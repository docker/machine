package provision

import (
	"bytes"
	"fmt"
)

func init() {
	Register("Ubuntu", &RegisteredProvisioner{
		New: NewUbuntuProvisioner,
	})
}

func NewUbuntuProvisioner(sshFunc SSHCommandFunc) Provisioner {
	return &UbuntuProvisioner{
		SSHCommand: sshFunc,
	}
}

type UbuntuProvisioner struct {
	OsReleaseInfo *OsRelease
	SSHCommand    SSHCommandFunc
}

func (provisioner *UbuntuProvisioner) Service(name string, action ServiceState) error {
	command := fmt.Sprintf("service %s %s", name, action.String())

	cmd, err := provisioner.SSHCommand(command)
	if err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) Package(name string, action PackageState) error {
	var packageState string

	switch action {
	case Installed:
		packageState = "install"
	case Missing:
		packageState = "remove"
	}

	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive sudo -E apt-get %s -y  %s", packageState, name)

	cmd, err := provisioner.SSHCommand(command)
	if err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) Hostname() (string, error) {
	cmd, err := provisioner.SSHCommand("hostname")
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

func (provisioner *UbuntuProvisioner) SetHostname(hostname string) error {
	cmd, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo \"%s\" | sudo tee /etc/hostname && echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts",
		hostname,
		hostname,
		hostname,
	))
	if err != nil {
		return err
	}

	return cmd.Run()
}

func (provisioner *UbuntuProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.Id == "Ubuntu"
}

func (provisioner *UbuntuProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}
