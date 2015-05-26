package provision

import (
	"fmt"
	"strconv"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
)

func newDebianFamilyProvisioner(d drivers.Driver, o string) DebianFamilyProvisioner {
	return DebianFamilyProvisioner{
		GenericProvisioner{
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/default/docker",
			OsReleaseId:       o,
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

type DebianFamilyProvisioner struct {
	GenericProvisioner
}

func (provisioner *DebianFamilyProvisioner) Service(name string, action pkgaction.ServiceAction) error {
	var command string
	if provisioner.GetInitSubsystem() == SYSTEMD {
		if _, err := provisioner.SSHCommand("sudo systemctl daemon-reload"); err != nil {
			return err
		}

		command = fmt.Sprintf("sudo systemctl %s %s", action.String(), name)
	} else {
		command = fmt.Sprintf("sudo service %s %s", name, action.String())
	}

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *DebianFamilyProvisioner) Package(name string, action pkgaction.PackageAction) error {
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
		if _, err := provisioner.SSHCommand("sudo -E apt-get update"); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive sudo -E apt-get %s -y  %s", packageAction, name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *DebianFamilyProvisioner) dockerDaemonResponding() bool {
	if _, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *DebianFamilyProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions, engineOptions engine.EngineOptions) error {
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

	if err := installDockerGeneric(provisioner); err != nil {
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

	if err := configureSwarm(provisioner, swarmOptions); err != nil {
		return err
	}

	return nil
}

func (provisioner *DebianFamilyProvisioner) GetInitSubsystem() int {
	if provisioner.OsReleaseId == "debian" {
		version, _ := strconv.Atoi(provisioner.OsReleaseInfo.VersionId)
		if version <= 7 {
			return SYSVINIT
		}
		provisioner.DaemonOptionsFile = "/lib/systemd/system/docker.service"
		return SYSTEMD
	} else if provisioner.OsReleaseId == "ubuntu" {
		return UPSTART
	}
	return SYSVINIT
}
