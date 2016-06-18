package provision

import (
	"fmt"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("Debian", &RegisteredProvisioner{
		New: NewDebianUpstartProvisioner,
	})
}

func NewDebianUpstartProvisioner(d drivers.Driver) Provisioner {
	return &DebianUpstartProvisioner{
		NewUpstartProvisioner("debian", d),
	}
}

type DebianUpstartProvisioner struct {
	UpstartProvisioner
}

func (provisioner *DebianUpstartProvisioner) String() string {
	return "debian(upstart)"
}

func (provisioner *DebianUpstartProvisioner) CompatibleWithHost() bool {
	isDebian := provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID
	if !isDebian {
		return false
	}
	if _, err := provisioner.SSHCommand("sudo which systemctl"); err != nil {
		return true
	}
	return false
}

func (provisioner *DebianUpstartProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	updateMetadata := true

	switch action {
	case pkgaction.Install, pkgaction.Upgrade:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "remove"
		updateMetadata = false
	}

	switch name {
	case "docker":
		name = "docker-engine"
	}

	if updateMetadata {
		if _, err := provisioner.SSHCommand("sudo apt-get update"); err != nil {
			return err
		}
	}

	// handle the new docker-engine package; we can probably remove this
	// after we have a few versions
	if action == pkgaction.Upgrade && name == "docker-engine" {
		// run the force remove on the existing lxc-docker package
		// and remove the existing apt source list

		commands := []string{
			"rm /etc/apt/sources.list.d/docker.list || true",
			"apt-get remove -y lxc-docker || true",
		}

		for _, cmd := range commands {
			command := fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive %s", cmd)
			if _, err := provisioner.SSHCommand(command); err != nil {
				return err
			}
		}
	}

	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive sudo -E apt-get %s -y  %s", packageAction, name)

	log.Debugf("package: action=%s name=%s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *DebianUpstartProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *DebianUpstartProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	storageDriver, err := decideStorageDriver(provisioner, "aufs", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	// HACK: since debian does not come with sudo by default we install
	log.Debug("installing sudo")
	if _, err := provisioner.SSHCommand("if ! type sudo; then apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y sudo; fi"); err != nil {
		return err
	}

	log.Debug("setting hostname")
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	log.Debug("installing base packages")
	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	log.Debug("installing docker")
	if err := installDockerGeneric(provisioner, engineOptions.InstallURL); err != nil {
		return err
	}

	log.Debug("waiting for docker daemon")
	if err := mcnutils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	log.Debug("configuring auth")
	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	log.Debug("configuring swarm")
	if err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions); err != nil {
		return err
	}

	return nil
}
