package awsec2

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/stripe/aws-go/aws"
	"github.com/stripe/aws-go/gen/ec2"
)

const (
	driverName = "awsec2"
)

func init() {
	drivers.Register(driverName, &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "aws-access-key-id",
			Usage:  "AWS Access Key",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "aws-secret-access-key",
			Usage:  "AWS Secret Access Key",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "aws-session-token",
			Usage:  "AWS Session Token",
			EnvVar: "AWS_SESSION_TOKEN",
		},
		cli.StringFlag{
			Name:   "aws-region",
			Usage:  "AWS Region",
			Value:  "eu-central-1",
			EnvVar: "AWS_DEFAULT_REGION",
		},

		cli.StringFlag{
			Name:  "aws-ec2-instance-type",
			Usage: "EC2 instance type",
			Value: "t2.small",
		},
		cli.StringFlag{
			Name:  "aws-ec2-image-id",
			Usage: "EC2 AMI ID",
		},
		cli.IntFlag{
			Name:  "aws-ec2-disk-size",
			Usage: "Size of the volume",
			Value: 16,
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{
		MachineName:    machineName,
		storePath:      storePath,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
	}, nil
}

type Driver struct {
	MachineName string

	AccessKey    string
	SecretKey    string
	SessionToken string
	Region       string

	InstanceID      string
	SecurityGroupID string
	KeyPath         string
	CaCertPath      string
	PrivateKeyPath  string

	ImageID      string
	InstanceType string
	DiskSize     int

	storePath string
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	log.Debug("Setting flags")
	d.AccessKey = flags.String("aws-access-key-id")
	d.SecretKey = flags.String("aws-secret-access-key")
	d.SessionToken = flags.String("aws-session-token")

	region, err := validateAwsRegion(flags.String("aws-region"))
	if err != nil {
		return err
	}

	d.Region = region

	image := flags.String("aws-ec2-image-id")
	if len(image) == 0 {
		image = regionDetails[d.Region].AmiId
	}

	d.InstanceType = flags.String("aws-ec2-instance-type")
	d.ImageID = image
	d.DiskSize = flags.Int("aws-ec2-disk-size")

	return nil
}

func (d *Driver) Create() error {
	resp, err := d.getClient().CreateKeyPair(&ec2.CreateKeyPairRequest{
		aws.Boolean(false),
		aws.String(d.MachineName),
	})

	if err != nil {
		return err
	}

	d.KeyPath = d.sshKeyPath()
	log.Debugf("Writing private SSH key to: %s", d.KeyPath)
	if err := ioutil.WriteFile(d.KeyPath, []byte(*resp.KeyMaterial), 0600); err != nil {
		return err
	}

	vpcId, err := d.getDefaultVpc()
	if err != nil {
		return err
	}

	log.Debugf("Creating security group")
	secGroupResp, err := d.getClient().CreateSecurityGroup(&ec2.CreateSecurityGroupRequest{
		Description: aws.String(fmt.Sprintf("Docker Machine host: %s", d.MachineName)),
		GroupName:   aws.String(d.MachineName),
		VPCID:       aws.String(vpcId),
	})

	if err != nil {
		return err
	}

	log.Debugf("Created security group: %s", *secGroupResp.GroupID)

	log.Debugf("Creating Docker Ingress rule")
	if err := d.getClient().AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressRequest{
		CIDRIP:     aws.String("0.0.0.0/0"),
		DryRun:     aws.Boolean(false),
		FromPort:   aws.Integer(2376),
		GroupID:    secGroupResp.GroupID,
		ToPort:     aws.Integer(2376),
		IPProtocol: aws.String("tcp"),
	}); err != nil {
		return err
	}

	log.Debugf("Creating SSH Ingress rule")
	if err := d.getClient().AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressRequest{
		CIDRIP:     aws.String("0.0.0.0/0"),
		DryRun:     aws.Boolean(false),
		FromPort:   aws.Integer(22),
		GroupID:    secGroupResp.GroupID,
		ToPort:     aws.Integer(22),
		IPProtocol: aws.String("tcp"),
	}); err != nil {
		return err
	}

	instResp, err := d.getClient().RunInstances(&ec2.RunInstancesRequest{
		BlockDeviceMappings: []ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				EBS: &ec2.EBSBlockDevice{
					DeleteOnTermination: aws.Boolean(false),
					VolumeSize:          aws.Integer(d.DiskSize),
					VolumeType:          aws.String("gp2"),
				},
			},
		},
		DryRun:           aws.Boolean(false),
		ImageID:          aws.String(d.ImageID),
		InstanceType:     aws.String(d.InstanceType),
		KeyName:          aws.String(d.MachineName),
		MinCount:         aws.Integer(1),
		MaxCount:         aws.Integer(1),
		SecurityGroupIDs: []string{*secGroupResp.GroupID},
	})

	if err != nil {
		return err
	}

	log.Debugf("Created instance: %s", *instResp.Instances[0].InstanceID)

	d.InstanceID = *instResp.Instances[0].InstanceID
	d.SecurityGroupID = *secGroupResp.GroupID

	d.getClient().CreateTags(&ec2.CreateTagsRequest{
		DryRun:    aws.Boolean(false),
		Resources: []string{d.InstanceID, d.SecurityGroupID},
		Tags: []ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(d.MachineName),
			},
		},
	})

	machineState, err := d.GetState()
	if err != nil {
		return err
	}

	if machineState != state.Running {
		return errMachineFailure
	}

	ip, err := d.GetIP()
	if err != nil {
		return err
	}

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", ip, 22)); err != nil {
		return err
	}

	log.Debugf("Setting hostname: %s", d.MachineName)
	cmd, err := d.GetSSHCommand(fmt.Sprintf(
		"echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts && sudo hostname %s && echo \"%s\" | sudo tee /etc/hostname",
		d.MachineName,
		d.MachineName,
		d.MachineName,
	))

	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Installing Docker")

	cmd, err = d.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl get.docker.io | sudo sh -; fi")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetDockerConfigDir() string {
	return "/etc/docker"
}

func (d *Driver) GetIP() (string, error) {
	descInstResp, err := d.getClient().DescribeInstances(&ec2.DescribeInstancesRequest{
		DryRun:      aws.Boolean(false),
		InstanceIDs: []string{d.InstanceID},
	})

	if err != nil {
		return "", err
	}

	if descInstResp.Reservations[0].Instances[0].PublicIPAddress == nil {
		return "unset", errNoIP
	}

	ip := *descInstResp.Reservations[0].Instances[0].PublicIPAddress
	log.Infof("Instance IP Address: %s", ip)

	return ip, nil
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ip, err := d.GetIP()
	if err != nil {
		return &exec.Cmd{}, err
	}

	return ssh.GetSSHCommand(ip, 22, "ubuntu", d.sshKeyPath(), args...), nil
}

func (d *Driver) GetState() (state.State, error) {
	for {
		log.Debugf("Polling instance status for %s", d.InstanceID)
		resp, err := d.getClient().DescribeInstances(&ec2.DescribeInstancesRequest{
			DryRun:      aws.Boolean(false),
			InstanceIDs: []string{d.InstanceID},
		})

		if err != nil {
			return state.Error, err
		}

		instance := resp.Reservations[0].Instances[0]
		instState := *instance.State.Name

		log.Debugf("Instance: state: %s", instState)

		if "running" == instState {
			return state.Running, nil
		}

		if "stopping" == instState || "shutting-down" == instState {
			return state.Stopping, nil
		}

		if "stopped" == instState {
			return state.Stopped, nil
		}

		if "terminated" == instState {
			return state.Stopped, nil
		}

		time.Sleep(1 * time.Second)
	}

	return state.None, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s:2376", ip), nil
}
func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Remove() error {
	_, err := d.getClient().TerminateInstances(&ec2.TerminateInstancesRequest{
		DryRun:      aws.Boolean(false),
		InstanceIDs: []string{d.InstanceID},
	})

	if err != nil {
		return err
	}

	for {
		instState, err := d.GetState()
		if err != nil {
			return nil
		}

		if instState == state.Stopped {
			break
		}

		if instState == state.Error {
			return errMachineFailure
		}
	}

	err = d.getClient().DeleteSecurityGroup(&ec2.DeleteSecurityGroupRequest{
		DryRun:  aws.Boolean(false),
		GroupID: aws.String(d.SecurityGroupID),
	})

	if err != nil {
		return err
	}

	err = d.getClient().DeleteKeyPair(&ec2.DeleteKeyPairRequest{
		DryRun:  aws.Boolean(false),
		KeyName: aws.String(d.MachineName),
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Restart() error {
	err := d.getClient().RebootInstances(&ec2.RebootInstancesRequest{
		DryRun:      aws.Boolean(false),
		InstanceIDs: []string{d.InstanceID},
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	_, err := d.getClient().StartInstances(&ec2.StartInstancesRequest{
		DryRun:      aws.Boolean(false),
		InstanceIDs: []string{d.InstanceID},
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	_, err := d.getClient().StopInstances(&ec2.StopInstancesRequest{
		DryRun:      aws.Boolean(false),
		InstanceIDs: []string{d.InstanceID},
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Upgrade() error {
	sshCmd, err := d.GetSSHCommand("apt-get update && apt-get install -y lxc-docker")
	if err != nil {
		return err
	}
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}

func (d *Driver) StartDocker() error {
	log.Debug("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker start")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) StopDocker() error {
	log.Debug("Stopping Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker stop")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) getClient() *ec2.EC2 {
	creds := aws.Creds(d.AccessKey, d.SecretKey, d.SessionToken)
	cli := ec2.New(creds, d.Region, nil)

	return cli
}

func (d *Driver) getDefaultVpc() (string, error) {
	log.Debugf("Trying to find default VPC for %s", d.Region)
	resp, err := d.getClient().DescribeVPCs(&ec2.DescribeVPCsRequest{
		DryRun: aws.Boolean(false),
		Filters: []ec2.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{"true"},
			},
		},
	})

	if err != nil {
		return "", err
	}

	if len(resp.VPCs) == 0 {
		return "", errNoVpcs
	}

	var vpcId string

	for _, v := range resp.VPCs {
		vpcId = *v.VPCID
	}

	log.Debugf("Found default VPC: %s", vpcId)

	return vpcId, nil
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}
