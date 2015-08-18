package fakedriver

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
)

type FakeDriver struct {
	*drivers.BaseDriver
	MockState state.State
	MockName  string
}

func (d *FakeDriver) DriverName() string {
	return "fakedriver"
}

func (d *FakeDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d *FakeDriver) GetURL() (string, error) {
	return "", nil
}

func (d *FakeDriver) GetMachineName() string {
	return d.MockName
}

func (d *FakeDriver) GetIP() (string, error) {
	return "1.2.3.4", nil
}

func (d *FakeDriver) GetSSHHostname() (string, error) {
	return "", nil
}

func (d *FakeDriver) GetSSHKeyPath() string {
	return ""
}

func (d *FakeDriver) GetSSHPort() (int, error) {
	return 0, nil
}

func (d *FakeDriver) GetSSHUsername() string {
	return ""
}

func (d *FakeDriver) GetState() (state.State, error) {
	return d.MockState, nil
}

func (d *FakeDriver) PreCreateCheck() error {
	return nil
}

func (d *FakeDriver) Create() error {
	return nil
}

func (d *FakeDriver) Remove() error {
	return nil
}

func (d *FakeDriver) Start() error {
	d.MockState = state.Running
	return nil
}

func (d *FakeDriver) Stop() error {
	d.MockState = state.Stopped
	return nil
}

func (d *FakeDriver) Restart() error {
	return nil
}

func (d *FakeDriver) Kill() error {
	return nil
}

func (d *FakeDriver) Upgrade() error {
	return nil
}
