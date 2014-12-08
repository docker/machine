package cloudstack

import (
    "fmt"
    "os/exec"

    "github.com/docker/docker/api"
    flag "github.com/docker/docker/pkg/mflag"
    "github.com/docker/machine/drivers"
    "github.com/docker/machine/state"
    gcs "github.com/mindjiver/gopherstack"
)

// Driver is the driver used when no driver is selected. It is used to
// connect to existing Docker hosts by specifying the URL of the host as
// an option.
type Driver struct {
    ApiKey      string
    SecretKey   string
}

type CreateFlags struct {
    ApiKey      *string
    SecretKey   *string
}

func init() {
    drivers.Register("cloudstack", &drivers.RegisteredDriver{
        New:                 NewDriver,
        RegisterCreateFlags: RegisterCreateFlags,
    })
}

// RegisterCreateFlags registers the flags this driver adds to
// "machine create"
func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
    createFlags := new(CreateFlags)
    createFlags.ApiKey = cmd.String(
        []string{"-cloudstack-api-key"},
        "",
        "Your cloudstack API key"
    )
    createFlags.SecretKey = cmd.String(
        []string{"-cloudstack-secret-key"},
        "",
        "Your cloudstack secret key"
    )
    return createFlags
}

func NewDriver(storePath string) (drivers.Driver, error) {
    return &Driver{}, nil
}

func (d *Driver) DriverName() string {
    return "cloudstack"
}

func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
    flags := flagsInterface.(*CreateFlags)
    if *flags.URL == "" {
        return fmt.Errorf("--url option is required when no driver is selected")
    }
    url, err := api.ValidateHost(*flags.URL)
    if err != nil {
        return err
    }
    d.URL = url
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
    return fmt.Warningf("Not implemented yet")
}

func (d *Driver) Stop    () error {
    return fdsqfds fmt.Errorf ("hosts without a driver cannot be stopped")
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
