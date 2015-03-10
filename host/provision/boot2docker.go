package provision

import (
	"bytes"
	"fmt"
	"regexp"
)

func init() {
	RegisterProvisioner("boot2docker", &ProvisionerFactories{
		New: NewBoot2DockerProvisioner,
	})
}

func NewBoot2DockerProvisioner(sshFunc SSHCommandFunc) Provisioner {
	return &Boot2DockerProvisioner{
		SSHCommandFunc: sshFunc,
	}
}

type Boot2DockerProvisioner struct {
	SSHCommandFunc SSHCommandFunc
}

func (provisioner *Boot2DockerProvisioner) Service(name string, action ServiceState) error {
	return nil
}

func (provisioner *Boot2DockerProvisioner) Package(name string, action PackageState) error {
	return nil
}

func (provisioner *Boot2DockerProvisioner) Hostname() (string, error) {
	cmd, err := provisioner.SSHCommandFunc(fmt.Sprintf("hostname"))
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
	cmd, err := provisioner.SSHCommandFunc(fmt.Sprintf(
		"sudo hostname %s && echo \"%s\" | sudo tee /var/lib/boot2docker/etc/hostname",
		hostname,
		hostname,
	))
	if err != nil {
		return err
	}

	return cmd.Run()
}

func (provisioner *Boot2DockerProvisioner) CompatibleWithHost() error {
	cmd, err := provisioner.SSHCommandFunc("cat /etc/os-release")
	if err != nil {
		return err
	}

	var so bytes.Buffer
	cmd.Stdout = &so

	if err := cmd.Run(); err != nil {
		return err
	}

	re := regexp.MustCompile(`(?m)^ID=(\w+)`)

	m := re.FindStringSubmatch(so.String())
	if len(m) > 0 && m[1] == "boot2docker" {
		return nil
	}

	return ErrDetectionFailed
}
