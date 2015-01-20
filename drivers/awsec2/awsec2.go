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
			EnvVar: "AWS_DEFAULT_REGION",
		},
	}
}

func NewDriver(machineName string, storePath string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath}, nil
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
			ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/sda1"),
				EBS: &ec2.EBSBlockDevice{
					VolumeSize: aws.Integer(8),
					VolumeType: aws.String("gp2"),
				},
			},
		},
		DryRun:           aws.Boolean(false),
		ImageID:          aws.String(regionDetails[d.Region].AmiId),
		InstanceType:     aws.String("t2.small"),
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

	cmd, err = d.GetSSHCommand("sudo stop docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("HACK: Downloading version of Docker with identity auth...")

	cmd, err = d.GetSSHCommand("sudo curl -sS -o /usr/bin/docker https://ehazlett.s3.amazonaws.com/public/docker/linux/docker-1.4.1-136b351e-identity")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Updating /etc/default/docker to use identity auth...")

	cmd, err = d.GetSSHCommand("echo 'export DOCKER_OPTS=\"--auth=identity --host=tcp://0.0.0.0:2376 --host=unix:///var/run/docker.sock --auth-authorized-dir=/root/.docker/authorized-keys.d\"' | sudo tee -a /etc/default/docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// HACK: create dir for ubuntu user to access
	log.Debugf("Adding key to authorized-keys.d...")

	cmd, err = d.GetSSHCommand("sudo mkdir -p /root/.docker && sudo chown -R ubuntu /root/.docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".docker/public-key.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	cmdString := fmt.Sprintf("sudo mkdir -p %q && sudo tee -a %q", "/root/.docker/authorized-keys.d", "/root/.docker/authorized-keys.d/docker-host.json")
	cmd, err = d.GetSSHCommand(cmdString)
	if err != nil {
		return err
	}
	cmd.Stdin = f
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand("sudo start docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
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
		return "", errNoIP
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

		// // This typically means that the instance doesn't exist or isn't booted yet
		// if insta.InstanceStatuses == nil {
		// 	time.Sleep(1 * time.Second)
		// 	continue
		// }

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
	return nil
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
			ec2.Filter{
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
