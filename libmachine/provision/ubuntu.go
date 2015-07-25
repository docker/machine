package provision

import (
	"fmt"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
)

func init() {
	Register("Ubuntu", &RegisteredProvisioner{
		New: NewUbuntuProvisioner,
	})
}

func NewUbuntuProvisioner(d drivers.Driver) Provisioner {
	return &UbuntuProvisioner{
		GenericProvisioner{
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/default/docker",
			OsReleaseId:       "ubuntu",
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

type UbuntuProvisioner struct {
	GenericProvisioner
}

func (provisioner *UbuntuProvisioner) Service(name string, action pkgaction.ServiceAction) error {
	sudo_service := provisioner.Driver.SSHSudo("service %s %s")
	command := fmt.Sprintf(sudo_service, name, action.String())

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var (
		packageAction  string
		updateMetadata = true
	)

	switch action {
	case pkgaction.Install:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
		updateMetadata = false
	case pkgaction.Upgrade:
		packageAction = "upgrade"
	}

	// TODO: This should probably have a const
	switch name {
	case "docker":
		name = "lxc-docker"
	}

	if updateMetadata {
		// issue apt-get update for metadata
		apt_update := provisioner.Driver.SSHSudo("apt-get update")
		if _, err := provisioner.SSHCommand(apt_update); err != nil {
			return err
		}
	}

	apt_action := provisioner.Driver.SSHSudo("apt-get %s -y %s")
	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive && %s", apt_action)
	command = fmt.Sprintf(command, packageAction, name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *UbuntuProvisioner) dockerDaemonResponding() bool {
	docker_version_command := provisioner.Driver.SSHSudo("docker version")
	if _, err := provisioner.SSHCommand(docker_version_command); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *UbuntuProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions, engineOptions engine.EngineOptions) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions

	if provisioner.EngineOptions.StorageDriver == "" {
		provisioner.EngineOptions.StorageDriver = "aufs"
	}

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	if err := installDockerGeneric(provisioner, engineOptions.InstallURL); err != nil {
		return err
	}

	if err := utils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	if err := makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	if err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions); err != nil {
		return err
	}

	return nil
}
