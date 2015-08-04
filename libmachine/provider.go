package libmachine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/utils"
)

type Provider struct {
	store Store
}

func New(store Store) (*Provider, error) {
	return &Provider{
		store: store,
	}, nil
}

func (provider *Provider) Create(name string, driverName string, hostOptions *HostOptions, driverConfig drivers.DriverOptions) (*Host, error) {
	validName := ValidateHostName(name)
	if !validName {
		return nil, ErrInvalidHostname
	}
	exists, err := provider.store.Exists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("Machine %s already exists", name)
	}

	hostPath := filepath.Join(utils.GetMachineDir(), name)

	host, err := NewHost(name, driverName, hostOptions)
	if err != nil {
		return host, err
	}
	if driverConfig != nil {
		if err := host.Driver.SetConfigFromFlags(driverConfig); err != nil {
			return host, err
		}
	}

	if err := host.Driver.PreCreateCheck(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return nil, err
	}

	if err := host.SaveConfig(); err != nil {
		return host, err
	}

	if err := host.Create(name); err != nil {
		return host, err
	}

	return host, nil
}

func (provider *Provider) Exists(name string) (bool, error) {
	return provider.store.Exists(name)
}

func (provider *Provider) GetActive() (*Host, error) {
	return provider.store.GetActive()
}

func (provider *Provider) List() ([]*Host, error) {
	return provider.store.List()
}

func (provider *Provider) Get(name string) (*Host, error) {
	return provider.store.Get(name)
}

func (provider *Provider) Remove(name string, force bool) error {
	host, err := provider.store.Get(name)
	if err != nil {
		return err
	}
	if err := host.Remove(force); err != nil {
		if !force {
			return err
		}
	}
	return provider.store.Remove(name, force)
}
