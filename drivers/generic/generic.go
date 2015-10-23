package generic

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	sourceSSHKey string
}

const (
	driverName     = "generic"
	defaultTimeout = 1 * time.Second
)

var (
	defaultSourceSSHKey = filepath.Join(mcnutils.GetHomeDir(), ".ssh", "id_rsa")
)

// NewDriver creates and returns a new instance of the driver
func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
		sourceSSHKey: defaultSourceSSHKey,
	}
}

func (d *Driver) Create() error {
	log.Info("Importing SSH key...")

	// TODO: validate the key is a valid key
	if err := mcnutils.CopyFile(d.sourceSSHKey, d.GetSSHKeyPath()); err != nil {
		return fmt.Errorf("unable to copy ssh key: %s", err)
	}

	if err := os.Chmod(d.GetSSHKeyPath(), 0600); err != nil {
		return fmt.Errorf("unable to set permissions on the ssh key: %s", err)
	}

	log.Debugf("IP: %s", d.IPAddress)

	return nil
}

func (d *Driver) DriverName() string {
	return driverName
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:  "generic-ip-address",
			Usage: "IP Address of machine",
		},
		mcnflag.StringFlag{
			Name:  "generic-ssh-user",
			Usage: "SSH user",
			Value: drivers.DefaultSSHUser,
		},
		mcnflag.StringFlag{
			Name:  "generic-ssh-key",
			Usage: "SSH private key path",
			Value: defaultSourceSSHKey,
		},
		mcnflag.IntFlag{
			Name:  "generic-ssh-port",
			Usage: "SSH port",
			Value: drivers.DefaultSSHPort,
		},
	}
}

func (d *Driver) GetState() (state.State, error) {
	address := net.JoinHostPort(d.IPAddress, strconv.Itoa(d.SSHPort))
	_, err := net.DialTimeout("tcp", address, defaultTimeout)
	var st state.State
	if err != nil {
		st = state.Stopped
	} else {
		st = state.Running
	}
	return st, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) Kill() error {
	log.Debug("Killing...")
	_, err := drivers.RunSSHCommandFromDriver(d, "sudo shutdown -P now")
	return err
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Restart() error {
	log.Debug("Restarting...")
	_, err := drivers.RunSSHCommandFromDriver(d, "sudo shutdown -r now")
	return err
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.IPAddress = flags.String("generic-ip-address")
	d.SSHUser = flags.String("generic-ssh-user")
	d.sourceSSHKey = flags.String("generic-ssh-key")
	d.SSHPort = flags.Int("generic-ssh-port")

	if d.IPAddress == "" {
		return errors.New("generic driver requires the --generic-ip-address option")
	}

	if d.sourceSSHKey == "" {
		return errors.New("generic driver requires the --generic-ssh-key option")
	}

	return nil
}

func (d *Driver) Start() error {
	return errors.New("generic driver does not support start")
}

func (d *Driver) Stop() error {
	return errors.New("generic driver does not support stop")
}
