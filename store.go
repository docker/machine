package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
)

// Store persists hosts on the filesystem
type Store struct {
	Path string
}

func NewStore(rootPath string) *Store {
	if rootPath == "" {
		rootPath = filepath.Join(drivers.GetHomeDir(), ".docker", "machines")
	}

	return &Store{Path: rootPath}
}

func (s *Store) Create(name string, driverName string, flags drivers.DriverOptions) (*Host, error) {
	exists, err := s.Exists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("Host %q already exists", name)
	}

	hostPath := filepath.Join(s.Path, name)

	host, err := NewHost(name, driverName, hostPath)
	if err != nil {
		return host, err
	}
	if flags != nil {
		if err := host.Driver.SetConfigFromFlags(flags); err != nil {
			return host, err
		}
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

func (s *Store) Remove(name string, force bool) error {
	active, err := s.GetActive()
	if err != nil {
		return err
	}
	if active != nil && active.Name == name {
		if err := s.RemoveActive(); err != nil {
			return err
		}
	}

	host, err := s.Load(name)
	if err != nil {
		return err
	}
	return host.Remove(force)
}

func (s *Store) List() ([]Host, error) {
	dir, err := ioutil.ReadDir(s.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	hosts := []Host{}

	for _, file := range dir {
		if file.IsDir() {
			host, err := s.Load(file.Name())
			if err != nil {
				log.Errorf("error loading host %q: %s", file.Name(), err)
				continue
			}
			hosts = append(hosts, *host)
		}
	}
	return hosts, nil
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

func (s *Store) Load(name string) (*Host, error) {
	hostPath := filepath.Join(s.Path, name)
	return LoadHost(name, hostPath)
}

func (s *Store) GetActive() (*Host, error) {
	hostName, err := ioutil.ReadFile(s.activePath())
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return s.Load(string(hostName))
}

func (s *Store) IsActive(host *Host) (bool, error) {
	active, err := s.GetActive()
	if err != nil {
		return false, err
	}
	if active == nil {
		return false, nil
	}
	return active.Name == host.Name, nil
}

func (s *Store) SetActive(host *Host) error {
	if err := os.MkdirAll(filepath.Dir(s.activePath()), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(s.activePath(), []byte(host.Name), 0600)
}

func (s *Store) RemoveActive() error {
	return os.Remove(s.activePath())
}

// activePath returns the path to the file that stores the name of the
// active host
func (s *Store) activePath() string {
	return filepath.Join(s.Path, ".active")
}
