package none

import (
	"fmt"
	"os/exec"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/api"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/state"
)

// Driver is the driver used when no driver is selected. It is used to
// connect to existing Docker hosts by specifying the URL of the host as
// an option.
type Driver struct {
	URL string
}

func init() {
	drivers.Register("none", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Usage: "URL of host when no driver is selected",
			Value: "",
		},
	}
}

func NewDriver(machineName string, storePath string) (drivers.Driver, error) {
	return &Driver{}, nil
}

func (d *Driver) DriverName() string {
	return "none"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	url := flags.String("url")

	if url == "" {
		return fmt.Errorf("--url option is required when no driver is selected")
	}
	validatedUrl, err := api.ValidateHost(url)
	if err != nil {
		return err
	}

	d.URL = validatedUrl
	return nil
}

func (d *Driver) GetURL() (string, error) {
	return d.URL, nil
}

func (d *Driver) GetIP() (string, error) {
	return "", nil
}

func (d *Driver) GetState() (state.State, error) {
	return state.None, nil
}

func (d *Driver) Create() error {
	return nil
}

func (d *Driver) Start() error {
	return fmt.Errorf("hosts without a driver cannot be started")
}

func (d *Driver) Stop() error {
	return fmt.Errorf("hosts without a driver cannot be stopped")
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Restart() error {
	return fmt.Errorf("hosts without a driver cannot be restarted")
}

func (d *Driver) Kill() error {
	return fmt.Errorf("hosts without a driver cannot be killed")
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("hosts without a driver cannot be upgraded")
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return nil, fmt.Errorf("hosts without a driver do not support SSH")
}
