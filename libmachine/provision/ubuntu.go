package provision

import (
	"bytes"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/utils"
)

func init() {
	Register("Ubuntu", &RegisteredProvisioner{
		New: NewUbuntuProvisioner,
	})
}

func NewUbuntuProvisioner(d drivers.Driver) Provisioner {
	return &UbuntuProvisioner{
		packages: []string{
			"curl",
		},
		Driver: d,
	}
}

type UbuntuProvisioner struct {
	packages      []string
	OsReleaseInfo *OsRelease
	Driver        drivers.Driver
	SwarmOptions  swarm.SwarmOptions
}

func (provisioner *UbuntuProvisioner) Service(name string, action pkgaction.ServiceAction) error {
	command := fmt.Sprintf("sudo service %s %s", name, action.String())

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	switch action {
	case pkgaction.Install:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
	case pkgaction.Upgrade:
		packageAction = "upgrade"
	}

	// TODO: This should probably have a const
	switch name {
	case "docker":
		name = "lxc-docker"
	}

	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive sudo -E apt-get %s -y  %s", packageAction, name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) dockerDaemonResponding() bool {
	if _, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warn("Error getting SSH command to check if the daemon is up: %s", err)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *UbuntuProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions) error {
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	for _, pkg := range provisioner.packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	if err := installDockerGeneric(provisioner); err != nil {
		return err
	}

	if err := utils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	if err := ConfigureAuth(provisioner, authOptions); err != nil {
		return err
	}

	if err := configureSwarm(provisioner, swarmOptions); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) Hostname() (string, error) {
	output, err := provisioner.SSHCommand("hostname")
	if err != nil {
		return "", err
	}

	var so bytes.Buffer
	if _, err := so.ReadFrom(output.Stdout); err != nil {
		return "", err
	}

	return so.String(), nil
}

func (provisioner *UbuntuProvisioner) SetHostname(hostname string) error {
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname && echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts",
		hostname,
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) GetDockerOptionsDir() string {
	return "/etc/docker"
}

func (provisioner *UbuntuProvisioner) SSHCommand(args string) (ssh.Output, error) {
	return drivers.RunSSHCommandFromDriver(provisioner.Driver, args)
}

func (provisioner *UbuntuProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.Id == "ubuntu"
}

func (provisioner *UbuntuProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *UbuntuProvisioner) GenerateDockerOptions(dockerPort int, authOptions auth.AuthOptions) (*DockerOptions, error) {
	defaultDaemonOpts := getDefaultDaemonOpts(provisioner.Driver.DriverName(), authOptions)
	daemonOpts := fmt.Sprintf("--host=unix:///var/run/docker.sock --host=tcp://0.0.0.0:%d", dockerPort)
	daemonOptsDir := "/etc/default/docker"
	opts := fmt.Sprintf("%s %s", defaultDaemonOpts, daemonOpts)
	daemonCfg := fmt.Sprintf("export DOCKER_OPTS=\\\"%s\\\"", opts)
	return &DockerOptions{
		EngineOptions:     daemonCfg,
		EngineOptionsPath: daemonOptsDir,
	}, nil
}

func (provisioner *UbuntuProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
