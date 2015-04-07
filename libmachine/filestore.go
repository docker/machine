package libmachine

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
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
		return nil, ErrHostDoesNotExist
	}

	host := &Host{Name: name, StorePath: hostPath}
	if err := host.LoadConfig(); err != nil {
		return nil, err
	}

	h := FillNestedHost(host)
	return h, nil
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
	hostName, err := ioutil.ReadFile(s.activePath())
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return s.Get(string(hostName))
}

func (s Filestore) IsActive(host *Host) (bool, error) {
	active, err := s.GetActive()
	if err != nil {
		return false, err
	}
	if active == nil {
		return false, nil
	}
	return active.Name == host.Name, nil
}

func (s Filestore) SetActive(host *Host) error {
	if err := os.MkdirAll(filepath.Dir(s.activePath()), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(s.activePath(), []byte(host.Name), 0600)
}

func (s Filestore) RemoveActive() error {
	return os.Remove(s.activePath())
}

// activePath returns the path to the file that stores the name of the
// active host
func (s Filestore) activePath() string {
	return filepath.Join(utils.GetMachineDir(), ".active")
}
