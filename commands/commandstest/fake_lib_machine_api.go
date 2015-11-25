package commandstest

import (
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/state"
)

type FakeLibmachineAPI struct {
	Hosts []*host.Host
}

func (api *FakeLibmachineAPI) NewPluginDriver(string, []byte) (drivers.Driver, error) {
	return nil, nil
}

func (api *FakeLibmachineAPI) NewHost(driver drivers.Driver) (*host.Host, error) {
	return nil, nil
}

func (api *FakeLibmachineAPI) Create(h *host.Host) error {
	return nil
}

func (api *FakeLibmachineAPI) Exists(name string) (bool, error) {
	for _, host := range api.Hosts {
		if name == host.Name {
			return true, nil
		}
	}

	return false, nil
}

func (api *FakeLibmachineAPI) List() ([]string, error) {
	return []string{}, nil
}

func (api *FakeLibmachineAPI) Load(name string) (*host.Host, error) {
	for _, host := range api.Hosts {
		if name == host.Name {
			return host, nil
		}
	}

	return nil, mcnerror.ErrHostDoesNotExist{
		Name: name,
	}
}

func (api *FakeLibmachineAPI) Remove(name string) error {
	newHosts := []*host.Host{}

	for _, host := range api.Hosts {
		if name != host.Name {
			newHosts = append(newHosts, host)
		}
	}

	api.Hosts = newHosts

	return nil
}

func (api *FakeLibmachineAPI) Save(host *host.Host) error {
	return nil
}

func State(api libmachine.API, name string) state.State {
	host, _ := api.Load(name)
	machineState, _ := host.Driver.GetState()
	return machineState
}

func Exists(api libmachine.API, name string) bool {
	exists, _ := api.Exists(name)
	return exists
}
