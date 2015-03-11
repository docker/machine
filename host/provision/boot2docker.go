package provision

import (
	"bytes"
	"fmt"
)

func init() {
	RegisterProvisioner("boot2docker", &ProvisionerFactories{
		New: NewBoot2DockerProvisioner,
	})
}

func NewBoot2DockerProvisioner(sshFunc SSHCommandFunc) Provisioner {
	return &Boot2DockerProvisioner{
		SSHCommand: sshFunc,
	}
}

type Boot2DockerProvisioner struct {
	OsReleaseInfo *OsRelease
	SSHCommand    SSHCommandFunc
}

func (provisioner *Boot2DockerProvisioner) Service(name string, action ServiceState) error {
	return nil
}

func (provisioner *Boot2DockerProvisioner) Package(name string, action PackageState) error {
	return nil
}

func (provisioner *Boot2DockerProvisioner) Hostname() (string, error) {
	cmd, err := provisioner.SSHCommand(fmt.Sprintf("hostname"))
	if err != nil {
		return "", err
	}

	var so bytes.Buffer
	cmd.Stdout = &so

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return so.String(), nil
}

func (provisioner *Boot2DockerProvisioner) SetHostname(hostname string) error {
	cmd, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo \"%s\" | sudo tee /var/lib/boot2docker/etc/hostname",
		hostname,
		hostname,
	))
	if err != nil {
		return err
	}

	return cmd.Run()
}

func (provisioner *Boot2DockerProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.Id == "boot2docker"
}

func (provisioner *Boot2DockerProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}
