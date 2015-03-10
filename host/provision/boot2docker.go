package provision

import (
	"github.com/docker/machine/drivers"
)

func init() {
	RegisterProvisioner("boot2docker", &ProvisionerDetection{
		New: func(driver drivers.Driver) Provisioner {
			return &Boot2DockerProvisioner{
				Driver: driver,
			}
		},
	})
}

type Boot2DockerProvisioner struct {
	Driver drivers.Driver
}

func (provisioner *Boot2DockerProvisioner) Service(name string, action ServiceState) error {
	return nil
}

func (provisioner *Boot2DockerProvisioner) Package(name string, action PackageState) error {
	return nil
}

func (provisioner *Boot2DockerProvisioner) Hostname() string {
	return ""
}

func (provisioner *Boot2DockerProvisioner) SetHostname(hostname string) error {
	return nil
}

func (provisioner *Boot2DockerProvisioner) CompatibleWithHost() error {
	return nil
}
