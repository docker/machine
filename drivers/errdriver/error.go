package errdriver

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	Name string
}

type ErrDriverNotLoadable struct {
	Name string
}

func (e ErrDriverNotLoadable) Error() string {
	return fmt.Sprintf("Driver %q not found. Do you have the plugin binary accessible in your PATH?", e.Name)
}

func NewDriver(Name string) drivers.Driver {
	return &Driver{
		Name: Name,
	}
}

func (d *Driver) DriverName() string {
	return "not-found"
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return ErrDriverNotLoadable{d.Name}
}

func (d *Driver) GetURL() (string, error) {
	return "", ErrDriverNotLoadable{d.Name}
}

func (d *Driver) GetMachineName() string {
	return ""
}

func (d *Driver) GetIP() (string, error) {
	return "1.2.3.4", ErrDriverNotLoadable{d.Name}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "", ErrDriverNotLoadable{d.Name}
}

func (d *Driver) GetSSHKeyPath() string {
	return ""
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, ErrDriverNotLoadable{d.Name}
}

func (d *Driver) GetSSHUsername() string {
	return ""
}

func (d *Driver) GetState() (state.State, error) {
	return state.Error, ErrDriverNotLoadable{d.Name}
}

func (d *Driver) Create() error {
	return ErrDriverNotLoadable{d.Name}
}

func (d *Driver) Remove() error {
	return ErrDriverNotLoadable{d.Name}
}

func (d *Driver) Start() error {
	return ErrDriverNotLoadable{d.Name}
}

func (d *Driver) Stop() error {
	return ErrDriverNotLoadable{d.Name}
}

func (d *Driver) Restart() error {
	return ErrDriverNotLoadable{d.Name}
}

func (d *Driver) Kill() error {
	return ErrDriverNotLoadable{d.Name}
}

func (d *Driver) Upgrade() error {
	return ErrDriverNotLoadable{d.Name}
}
