package api

import (
	"io/ioutil"
	"os"
	"path/filepath"

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
	return LoadMachine(name, machinePath)
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
