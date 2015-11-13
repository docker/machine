package rpcdriver

import (
	"fmt"

	"encoding/json"

	"github.com/docker/machine/drivers/errdriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
)

type Store interface {
	libmachine.HostSaver

	// NewHost will initialize a new host machine
	NewHost(driverName string, machineName string) (*host.Host, error)

	// List returns a list of all hosts in the store
	List() ([]*host.Host, error)

	// Load loads a host by name
	Load(name string) (*host.Host, error)

	// Remove removes a machine from the store
	Remove(name string) error
}

type rpcStore struct {
	*persist.Filestore
}

func NewRPCStore(fileStore *persist.Filestore) Store {
	return &rpcStore{fileStore}
}

// Exists returns whether a machine exists or not
func (s *rpcStore) Exists(name string) (bool, error) {
	return s.Filestore.Exists(name)
}

// NewHost will initialize a new host machine
func (s *rpcStore) NewHost(driverName string, machineName string) (*host.Host, error) {
	// TODO: Fix hacky JSON solution
	rawContent, err := json.Marshal(&drivers.BaseDriver{
		MachineName: machineName,
		StorePath:   s.Filestore.Path,
	})
	if err != nil {
		return nil, fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	driver, err := newPluginDriver(driverName, rawContent)
	if err != nil {
		return nil, err
	}

	return s.Filestore.NewHost(driver)
}

// List returns a list of all hosts in the store
func (s *rpcStore) List() ([]*host.Host, error) {
	cliHosts := []*host.Host{}

	hosts, err := s.Filestore.List()
	if err != nil {
		return nil, fmt.Errorf("Error attempting to list hosts from store: %s", err)
	}

	for _, h := range hosts {
		driver, err := newPluginDriver(h.DriverName, h.RawDriver)
		if err != nil {
			return nil, err
		}

		h.Driver = driver

		cliHosts = append(cliHosts, h)
	}

	return cliHosts, nil
}

// Load loads a host by name
func (s *rpcStore) Load(name string) (*host.Host, error) {
	h, err := s.Filestore.Load(name)
	if err != nil {
		return nil, fmt.Errorf("Loading host from store failed: %s", err)
	}

	driver, err := newPluginDriver(h.DriverName, h.RawDriver)
	if err != nil {
		return nil, err
	}

	h.Driver = driver

	return h, nil
}

// Remove removes a machine from the store
func (s *rpcStore) Remove(name string) error {
	return s.Filestore.Remove(name)
}

// Save persists a machine in the store
func (s *rpcStore) Save(host *host.Host) error {
	if serialDriver, ok := host.Driver.(*drivers.SerialDriver); ok {
		// Unwrap Driver
		host.Driver = serialDriver.Driver

		// Re-wrap Driver when done
		defer func() {
			host.Driver = serialDriver
		}()
	}

	if rpcClientDriver, ok := host.Driver.(*RPCClientDriver); ok {
		data, err := rpcClientDriver.GetConfigRaw()
		if err != nil {
			return persist.NewSaveError(host.Name, fmt.Errorf("Unable to get raw config for driver: %s", err))
		}
		host.RawDriver = data
	}

	return s.Filestore.Save(host)
}

func newPluginDriver(driverName string, rawContent []byte) (drivers.Driver, error) {
	d, err := NewRPCClientDriver(rawContent, driverName)
	if err != nil {
		// Not being able to find a driver binary is a "known error"
		if _, ok := err.(localbinary.ErrPluginBinaryNotFound); ok {
			return errdriver.NewDriver(driverName), nil
		}
		return nil, fmt.Errorf("Error attempting to invoke binary for plugin %q: %s", driverName, err)
	}

	if driverName == "virtualbox" {
		return drivers.NewSerialDriver(d), nil
	}

	return d, nil
}
