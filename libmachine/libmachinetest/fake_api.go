package libmachinetest

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist/persisttest"
)

type FakeAPI struct {
	*persisttest.FakeStore
	FakeNewDriver                             drivers.Driver
	FakeNewHost                               *host.Host
	NewPluginDriverErr, NewHostErr, CreateErr error
}

func (fapi *FakeAPI) NewPluginDriver(string, []byte) (drivers.Driver, error) {
	return fapi.FakeNewDriver, fapi.NewPluginDriverErr
}

func (fapi *FakeAPI) NewHost(drivers.Driver) (*host.Host, error) {
	return fapi.FakeNewHost, fapi.NewHostErr
}

func (fapi *FakeAPI) Create(h *host.Host) error {
	return fapi.CreateErr
}
