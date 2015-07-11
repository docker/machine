package libmachine

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
)

type Filestore struct {
	path           string
	caCertPath     string
	privateKeyPath string
}

func NewFilestore(rootPath string, caCert string, privateKey string) *Filestore {
	return &Filestore{path: rootPath, caCertPath: caCert, privateKeyPath: privateKey}
}

func (s Filestore) loadHost(name string) (*Host, error) {
	hostPath := filepath.Join(utils.GetMachineDir(), name)
	if _, err := os.Stat(hostPath); os.IsNotExist(err) {
		return nil, ErrHostDoesNotExist{
			Name: name,
		}
	}

	host := &Host{Name: name, StorePath: hostPath}
	if err := host.LoadConfig(); err != nil {
		return nil, err
	}

	return host, nil
}

func (s Filestore) GetPath() string {
	return s.path
}

func (s Filestore) GetCACertificatePath() (string, error) {
	return s.caCertPath, nil
}

func (s Filestore) GetPrivateKeyPath() (string, error) {
	return s.privateKeyPath, nil
}

func (s Filestore) Save(host *Host) error {
	data, err := json.Marshal(host)
	if err != nil {
		return err
	}

	hostPath := filepath.Join(utils.GetMachineDir(), host.Name)

	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(hostPath, "config.json"), data, 0600); err != nil {
		return err
	}

	return nil
}

func (s Filestore) Remove(name string, force bool) error {
	hostPath := filepath.Join(utils.GetMachineDir(), name)
	return os.RemoveAll(hostPath)
}

func (s Filestore) List() ([]*Host, error) {
	dir, err := ioutil.ReadDir(utils.GetMachineDir())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	hosts := []*Host{}

	for _, file := range dir {
		// don't load hidden dirs; used for configs
		if file.IsDir() && strings.Index(file.Name(), ".") != 0 {
			host, err := s.Get(file.Name())
			if err != nil {
				log.Errorf("error loading host %q: %s", file.Name(), err)
				continue
			}
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

func (s Filestore) Exists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(utils.GetMachineDir(), name))

	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}

	return false, err
}

func (s Filestore) Get(name string) (*Host, error) {
	return s.loadHost(name)
}

func (s Filestore) GetActive() (*Host, error) {
	hosts, err := s.List()
	if err != nil {
		return nil, err
	}

	dockerHost := os.Getenv("DOCKER_HOST")
	hostListItems := GetHostListItems(hosts)

	for _, item := range hostListItems {
		if dockerHost == item.URL {
			host, err := s.Get(item.Name)
			if err != nil {
				return nil, err
			}
			return host, nil
		}
	}

	return nil, errors.New("Active host not found")
}
