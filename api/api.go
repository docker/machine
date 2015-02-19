package api

import (
	"os"
	"path/filepath"

	"github.com/docker/machine"
	"github.com/docker/machine/drivers"
)

type Api struct {
	store *machine.Store
}

func NewApi(storePath, caCertPath, privateKeyPath string) (*Api, error) {
	st := machine.NewStore(storePath, caCertPath, privateKeyPath)
	api := &Api{
		store: st,
	}
	return api, nil
}

func (m *Api) Create(name string, driverName string, flags drivers.DriverOptions) (*machine.Machine, error) {
	exists, err := m.store.Exists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, machine.ErrMachineExists
	}

	machinePath := filepath.Join(m.store.Path, name)

	machine, err := machine.NewMachine(name, driverName, machinePath, m.store.CaCertPath, m.store.PrivateKeyPath)
	if err != nil {
		return machine, err
	}
	if flags != nil {
		if err := machine.Driver.SetConfigFromFlags(flags); err != nil {
			return machine, err
		}
	}

	if err := machine.Driver.PreCreateCheck(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(machinePath, 0700); err != nil {
		return nil, err
	}

	if err := m.store.Save(machine); err != nil {
		return machine, err
	}

	if err := machine.Create(name); err != nil {
		return machine, err
	}

	if err := machine.ConfigureAuth(); err != nil {
		return machine, err
	}

	return machine, nil
}

func (m *Api) Remove(name string, force bool) error {
	active, err := m.store.GetActive()
	if err != nil {
		return err
	}

	if active != nil && active.Name == name {
		if err := m.store.RemoveActive(); err != nil {
			return err
		}
	}

	machine, err := m.Get(name)
	if err != nil {
		return err
	}

	return machine.Remove(force)
}

func (m *Api) List() ([]machine.Machine, []error) {
	return m.store.List()
}

func (m *Api) Exists(name string) (bool, error) {
	return m.store.Exists(name)
}

func (m *Api) Get(name string) (*machine.Machine, error) {
	return m.store.Get(name)
}

func (m *Api) GetActive() (*machine.Machine, error) {
	return m.store.GetActive()
}

func (m *Api) IsActive(mcn *machine.Machine) (bool, error) {
	return m.store.IsActive(mcn)
}

func (m *Api) SetActive(mcn *machine.Machine) error {
	return m.store.SetActive(mcn)
}

func (m *Api) RemoveActive() error {
	return m.store.RemoveActive()
}

func (m *Api) Start(name string) error {
	return nil
}

func (m *Api) Stop(name string) error {
	return nil
}

func (m *Api) Restart(name string) error {
	return nil
}

func (m *Api) Kill(name string) error {
	return nil
}

func (m *Api) Upgrade(name string) error {
	return nil
}
