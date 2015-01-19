package awsec2

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/docker/machine/drivers"
	// "github.com/docker/machine/ssh"
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

	KeyPath string

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

	d.KeyPath = filepath.Join(d.storePath, "id_rsa")
	log.Debugf("Writing private SSH key to: %s", d.KeyPath)
	if err := ioutil.WriteFile(d.KeyPath, []byte(*resp.KeyMaterial), 0600); err != nil {
		return err
	}

	vpcId, err := d.getDefaultVpc()
	if err != nil {
		return err
	}

	log.Debugf("Creating security group")
	_, err = d.getClient().CreateSecurityGroup(&ec2.CreateSecurityGroupRequest{
		Description: aws.String(fmt.Sprintf("Docker Machine host: %s", d.MachineName)),
		GroupName:   aws.String(d.MachineName),
		VPCID:       aws.String(vpcId),
	})

	if err != nil {
		return err
	}

	return errComplete
}

func (d *Driver) GetIP() (string, error) {
	return "127.0.0.1", nil
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return &exec.Cmd{}, nil
}

func (d *Driver) GetState() (state.State, error) {
	return state.None, nil
}

func (d *Driver) GetURL() (string, error) {
	return "tcp://localhost:2376", nil
}
func (d *Driver) Kill() error {
	return nil
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Restart() error {
	return nil
}

func (d *Driver) Start() error {
	return nil
}

func (d *Driver) Stop() error {
	return nil
}

func (d *Driver) Upgrade() error {
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
