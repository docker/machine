package api

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/machine/drivers"
)

type Api struct {
	store *Store
}

func NewApi(storePath, caCertPath, privateKeyPath string) (*Api, error) {
	st := NewStore(storePath, caCertPath, privateKeyPath)
	api := &Api{
		store: st,
	}
	return api, nil
}

func (m *Api) Create(name string, driverName string, flags drivers.DriverOptions) (*Machine, error) {
	exists, err := m.store.Exists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrMachineExists
	}

	machinePath := filepath.Join(m.store.Path, name)

	machine, err := NewMachine(name, driverName, machinePath, m.store.CaCertPath, m.store.PrivateKeyPath)
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

	if err := machine.SaveConfig(); err != nil {
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

func (m *Api) List() ([]Machine, []error) {
	dir, err := ioutil.ReadDir(m.store.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, []error{err}
	}

	machines := []Machine{}
	errors := []error{}

	for _, file := range dir {
		// don't load hidden dirs; used for configs
		if file.IsDir() && strings.Index(file.Name(), ".") != 0 {
			machine, err := m.Get(file.Name())
			if err != nil {
				errors = append(errors, err)
				continue
			}
			machines = append(machines, *machine)
		}
	}
	return machines, nil
}

func (m *Api) Exists(name string) (bool, error) {
	return m.store.Exists(name)
}

func (m *Api) Get(name string) (*Machine, error) {
	machinePath := filepath.Join(m.store.Path, name)
	return LoadMachine(name, machinePath)
}

func (m *Api) GetActive() (*Machine, error) {
	return m.store.GetActive()
}

func (m *Api) IsActive(machine *Machine) (bool, error) {
	return m.store.IsActive(machine)
}

func (m *Api) SetActive(machine *Machine) error {
	return m.store.SetActive(machine)
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
