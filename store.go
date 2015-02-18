package machine

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/machine/utils"
)

// Store persists hosts on the filesystem
type Store struct {
	Path           string
	CaCertPath     string
	PrivateKeyPath string
}

func NewStore(rootPath string, caCert string, privateKey string) *Store {
	if rootPath == "" {
		rootPath = utils.GetMachineDir()
	}

	return &Store{Path: rootPath, CaCertPath: caCert, PrivateKeyPath: privateKey}
}

func loadMachine(name string, storePath string) (*Machine, error) {
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return nil, ErrMachineDoesNotExist
	}

	machine := &Machine{Name: name, StorePath: storePath}
	if err := machine.LoadConfig(); err != nil {
		return nil, err
	}

	return machine, nil
}

func (s *Store) Save(m *Machine) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(m.StorePath, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(m.StorePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}

func (s *Store) Exists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.Path, name))
	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

func (s *Store) Get(name string) (*Machine, error) {
	machinePath := filepath.Join(s.Path, name)
	return loadMachine(name, machinePath)
}

func (s *Store) List() ([]Machine, []error) {
	dir, err := ioutil.ReadDir(s.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, []error{err}
	}

	machines := []Machine{}
	errors := []error{}

	for _, file := range dir {
		// don't load hidden dirs; used for configs
		if file.IsDir() && strings.Index(file.Name(), ".") != 0 {
			machine, err := s.Get(file.Name())
			if err != nil {
				errors = append(errors, err)
				continue
			}
			machines = append(machines, *machine)
		}
	}
	return machines, nil
}

func (s *Store) GetActive() (*Machine, error) {
	machineName, err := ioutil.ReadFile(s.activePath())
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return s.Get(string(machineName))
}

func (s *Store) IsActive(machine *Machine) (bool, error) {
	active, err := s.GetActive()
	if err != nil {
		return false, err
	}
	if active == nil {
		return false, nil
	}
	return active.Name == machine.Name, nil
}

func (s *Store) SetActive(machine *Machine) error {
	if err := os.MkdirAll(filepath.Dir(s.activePath()), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(s.activePath(), []byte(machine.Name), 0600)
}

func (s *Store) RemoveActive() error {
	return os.Remove(s.activePath())
}

// activePath returns the path to the file that stores the name of the
// active host
func (s *Store) activePath() string {
	return filepath.Join(s.Path, ".active")
}
