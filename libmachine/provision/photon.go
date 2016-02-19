package provision

import (
	"fmt"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
)

func init() {
	Register("Photon", &RegisteredProvisioner{
		New: NewPhotonProvisioner,
	})
}

func NewPhotonProvisioner(d drivers.Driver) Provisioner {
	return &PhotonProvisioner{
		NewSystemdProvisioner("photon", d),
	}
}

type PhotonProvisioner struct {
	SystemdProvisioner
}

func (provisioner *PhotonProvisioner) Package(name string, action pkgaction.PackageAction) error {
	var (
		packageAction  string
		updateMetadata = true
	)

	switch action {
	case pkgaction.Install:
		packageAction = "install"
	case pkgaction.Remove:
		packageAction = "erase"
		updateMetadata = false
	case pkgaction.Upgrade:
		packageAction = "upgrade"
	}

	if updateMetadata {
		if _, err := provisioner.SSHCommand("sudo -E tdnf --assumeyes makecache"); err != nil {
			return err
		}
	}
	// Please look away :-)
	command := fmt.Sprintf("sudo -E tdnf --assumeyes %s %s; if [ $? -eq 137 ]; then echo 'already installed'; fi", packageAction, name)

	if output, err := provisioner.SSHCommand(command); err != nil {
		return fmt.Errorf(output)
	}

	return nil
}

func (provisioner *PhotonProvisioner) dockerDaemonResponding() bool {
	if _, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *PhotonProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	if provisioner.EngineOptions.StorageDriver == "" {
		provisioner.EngineOptions.StorageDriver = "overlay"
	} else if provisioner.EngineOptions.StorageDriver != "overlay" {
		return fmt.Errorf("Unsupported storage driver: %s", provisioner.EngineOptions.StorageDriver)
	}

	log.Debugf("Setting hostname %s", provisioner.Driver.GetMachineName())
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
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

	if err := waitForDocker(provisioner, 2376); err != nil {
		return err
	}

	return nil
}
