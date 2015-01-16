package awsec2

import (
	"os/exec"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/state"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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

	storePath string
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d *Driver) Create() error {
	return nil
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
