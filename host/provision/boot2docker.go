package provision

import (
	"bytes"
	"regexp"

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
	cmd, err := provisioner.Driver.GetSSHCommand("cat /etc/os-release")
	if err != nil {
		return err
	}

	var so bytes.Buffer
	cmd.Stdout = &so

	if err := cmd.Run(); err != nil {
		return err
	}

	re := regexp.MustCompile(`(?m)^ID=(\w+)`)

	if len(re.FindStringSubmatch(so.String())) > 0 {
		return nil
	}

	return ErrDetectionFailed
}
