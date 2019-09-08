package provision

import (
	"fmt"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("Alpine", &RegisteredProvisioner{
		New: NewAlpineProvisioner,
	})
}

func NewAlpineProvisioner(d drivers.Driver) Provisioner {
	return &AlpineProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/conf.d/docker",
			OsReleaseID:       "alpine",
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

type AlpineProvisioner struct {
	GenericProvisioner
}

func (provisioner *AlpineProvisioner) String() string {
	return "alpine"
}

func (provisioner *AlpineProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID || provisioner.OsReleaseInfo.IDLike == provisioner.OsReleaseID
}

func (provisioner *AlpineProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var packageAction string

	switch action {
	case pkgaction.Install, pkgaction.Upgrade:
		packageAction = "add"
	case pkgaction.Remove:
		packageAction = "remove"
	}

	switch name {
	case "docker-engine":
		name = "docker"
	case "docker":
		name = "docker"
	}

	command := fmt.Sprintf("sudo apk %s %s", packageAction, name)

	log.Debugf("package: action=%s name=%s", action.String(), name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *AlpineProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *AlpineProvisioner) Service(name string, action serviceaction.ServiceAction) error {	
	var command string
	switch action {
	case serviceaction.Start, serviceaction.Restart, serviceaction.Stop:
		command = fmt.Sprintf("sudo rc-service %s %s", name, action.String())
	case serviceaction.DaemonReload:
		command = fmt.Sprintf("sudo rc-service restart %s", name)
	case serviceaction.Enable:
		command = fmt.Sprintf("sudo rc-update add %s boot", name)
	case serviceaction.Disable:
		command = fmt.Sprintf("sudo rc-update del %s boot", name)
	}
	
	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *AlpineProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	storageDriver, err := decideStorageDriver(provisioner, "overlay", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	// HACK: since Alpine does not come with sudo by default we install
	log.Debug("Installing sudo")
	if _, err := provisioner.SSHCommand("if ! type sudo; then apk add sudo; fi"); err != nil {
		return err
	}

	log.Debug("Setting hostname")
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	log.Debug("Installing base packages")
	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	log.Debug("Installing docker")
	if err := provisioner.Package("docker", pkgaction.Install); err != nil {
		return err
	}

	log.Debug("Starting docker service")
	if err := provisioner.Service("docker", serviceaction.Start); err != nil {
		return err
	}

	log.Debug("Waiting for docker daemon")
	if err := mcnutils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	log.Debug("Configuring auth")
	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	log.Debug("Configuring swarm")
	if err := configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions); err != nil {
		return err
	}

	// enable service
	log.Debug("Enabling docker service")
	if err := provisioner.Service("docker", serviceaction.Enable); err != nil {
		return err
	}

	return nil
}