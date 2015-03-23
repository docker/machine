package libmachine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/utils"
)

type Machine struct {
	store Store
}

func New(store Store) (*Machine, error) {
	return &Machine{
		store: store,
	}, nil
}

func (m *Machine) Create(name string, driverName string, hostOptions *HostOptions, driverConfig drivers.DriverOptions) (*Host, error) {
	validName := ValidateHostName(name)
	if !validName {
		return nil, ErrInvalidHostname
	}
	exists, err := m.store.Exists(name)
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

	if err := m.store.SetActive(host); err != nil {
		return nil, err
	}

	return host, nil
}

func (m *Machine) Exists(name string) (bool, error) {
	return m.store.Exists(name)
}

func (m *Machine) GetActive() (*Host, error) {
	return m.store.GetActive()
}

func (m *Machine) IsActive(host *Host) (bool, error) {
	return m.store.IsActive(host)
}

func (m *Machine) List() ([]*Host, error) {
	return m.store.List()
}

func (m *Machine) Get(name string) (*Host, error) {
	return m.store.Get(name)
}

func (m *Machine) Remove(name string, force bool) error {
	active, err := m.store.GetActive()
	if err != nil {
		return err
	}

	if active != nil && active.Name == name {
		if err := m.RemoveActive(); err != nil {
			return err
		}
	}

	host, err := m.store.Get(name)
	if err != nil {
		return err
	}
	if err := host.Remove(force); err != nil {
		if !force {
			return err
		}
	}
	return m.store.Remove(name, force)
}

func (m *Machine) RemoveActive() error {
	return m.store.RemoveActive()
}

func (m *Machine) SetActive(host *Host) error {
	return m.store.SetActive(host)
}
