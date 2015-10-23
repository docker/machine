package fakedriver

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	MockState state.State
	MockURL   string
	MockName  string
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{}
}

func (d *Driver) DriverName() string {
	return "Driver"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d *Driver) GetURL() (string, error) {
	return d.MockURL, nil
}

func (d *Driver) GetMachineName() string {
	return d.MockName
}

func (d *Driver) GetIP() (string, error) {
	return "1.2.3.4", nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "", nil
}

func (d *Driver) GetSSHKeyPath() string {
	return ""
}

func (d *Driver) GetSSHPort() (int, error) {
	return 0, nil
}

func (d *Driver) GetSSHUsername() string {
	return ""
}

func (d *Driver) GetState() (state.State, error) {
	return d.MockState, nil
}

func (d *Driver) Create() error {
	return nil
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Start() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) Stop() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) Restart() error {
	return nil
}

func (d *Driver) Kill() error {
	return nil
}

func (d *Driver) Upgrade() error {
	return nil
}
