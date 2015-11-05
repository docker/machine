package persist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/machine/drivers/errdriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/version"
)

type Filestore struct {
	Path             string
	CaCertPath       string
	CaPrivateKeyPath string
}

func (s Filestore) getMachinesDir() string {
	return filepath.Join(s.Path, "machines")
}

func (s Filestore) saveToFile(data []byte, file string) error {
	return ioutil.WriteFile(file, data, 0600)
}

func (s Filestore) Save(host *host.Host) error {
	// TODO: Does this belong here?
	if rpcClientDriver, ok := host.Driver.(*rpcdriver.RpcClientDriver); ok {
		data, err := rpcClientDriver.GetConfigRaw()
		if err != nil {
			return fmt.Errorf("Error getting raw config for driver: %s", err)
		}
		host.RawDriver = data
	}

	data, err := json.MarshalIndent(host, "", "    ")
	if err != nil {
		return err
	}

	hostPath := filepath.Join(s.getMachinesDir(), host.Name)

	// Ensure that the directory we want to save to exists.
	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return err
	}

	return s.saveToFile(data, filepath.Join(hostPath, "config.json"))
}

func (s Filestore) Remove(name string) error {
	hostPath := filepath.Join(s.getMachinesDir(), name)
	return os.RemoveAll(hostPath)
}

func (s Filestore) List() ([]*host.Host, error) {
	dir, err := ioutil.ReadDir(s.getMachinesDir())
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	hosts := []*host.Host{}

	for _, file := range dir {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			h, err := s.Load(file.Name())
			if err != nil {
				log.Errorf("error loading host %q: %s", file.Name(), err)
				continue
			}

			d, err := newPluginDriver(h.DriverName, h.RawDriver)
			if err != nil {
				return nil, err
			}

			h.Driver = d

			hosts = append(hosts, h)
		}
	}

	return hosts, nil
}

func (s Filestore) Exists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.getMachinesDir(), name))

	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}

	return false, err
}

func (s Filestore) loadConfig(h *host.Host) error {
	data, err := ioutil.ReadFile(filepath.Join(s.getMachinesDir(), h.Name, "config.json"))
	if err != nil {
		return err
	}

	// Remember the machine name so we don't have to pass it through each
	// struct in the migration.
	name := h.Name

	migratedHost, migrationPerformed, err := host.MigrateHost(h, data)
	if err != nil {
		return fmt.Errorf("Error getting migrated host: %s", err)
	}

	*h = *migratedHost

	h.Name = name

	// If we end up performing a migration, we should save afterwards so we don't have to do it again on subsequent invocations.
	if migrationPerformed {
		if err := s.saveToFile(data, filepath.Join(s.getMachinesDir(), h.Name, "config.json.bak")); err != nil {
			return fmt.Errorf("Error attempting to save backup after migration: %s", err)
		}

		if err := s.Save(h); err != nil {
			return fmt.Errorf("Error saving config after migration was performed: %s", err)
		}
	}

	return nil

}

func (s Filestore) Load(name string) (*host.Host, error) {
	hostPath := filepath.Join(s.getMachinesDir(), name)

	_, err := os.Stat(hostPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, mcnerror.ErrHostDoesNotExist{
				Name: name,
			}
		} else {
			return nil, fmt.Errorf("Error checking if host %q exists: %s", name, err)
		}
	}

	h := &host.Host{
		Name: name,
	}

	if err := s.loadConfig(h); err != nil {
		return nil, err
	}

	d, err := newPluginDriver(h.DriverName, h.RawDriver)
	if err != nil {
		return nil, err
	}

	h.Driver = d

	return h, nil
}

func (s Filestore) NewHost(name string, driverName string) (*host.Host, error) {
	// TODO: Fix hacky JSON solution
	rawContent, err := json.Marshal(&drivers.BaseDriver{
		MachineName: name,
		StorePath:   s.Path,
	})
	if err != nil {
		return nil, fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	driver, err := newPluginDriver(driverName, rawContent)
	if err != nil {
		return nil, err
	}

	certDir := filepath.Join(s.Path, "certs")

	hostOptions := &host.HostOptions{
		AuthOptions: &auth.AuthOptions{
			CertDir:          certDir,
			CaCertPath:       filepath.Join(certDir, "ca.pem"),
			CaPrivateKeyPath: filepath.Join(certDir, "ca-key.pem"),
			ClientCertPath:   filepath.Join(certDir, "cert.pem"),
			ClientKeyPath:    filepath.Join(certDir, "key.pem"),
			ServerCertPath:   filepath.Join(s.getMachinesDir(), "server.pem"),
			ServerKeyPath:    filepath.Join(s.getMachinesDir(), "server-key.pem"),
		},
		EngineOptions: &engine.EngineOptions{
			InstallURL:    "https://get.docker.com",
			StorageDriver: "aufs",
			TlsVerify:     true,
		},
		SwarmOptions: &swarm.SwarmOptions{
			Host:     "tcp://0.0.0.0:3376",
			Image:    "swarm:latest",
			Strategy: "spread",
		},
	}

	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          name,
		Driver:        driver,
		DriverName:    driverName,
		HostOptions:   hostOptions,
	}, nil
}

func newPluginDriver(driverName string, rawContent []byte) (drivers.Driver, error) {
	d, err := rpcdriver.NewRpcClientDriver(rawContent, driverName)
	if err != nil {
		// Not being able to find a driver binary is a "known error"
		if _, ok := err.(localbinary.ErrPluginBinaryNotFound); ok {
			return errdriver.NewDriver(driverName), nil
		}
		return nil, fmt.Errorf("Error loading driver %q: %s", driverName, err)
	}

	return d, nil
}
