package provision

import (
	"errors"
	"fmt"
	"strconv"

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
	Register("Poky", &RegisteredProvisioner{
		New: NewPokyProvisioner,
	})
}

func NewPokyProvisioner(d drivers.Driver) Provisioner {
	return &PokyProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/default/docker",
			OsReleaseID:       "poky",
			Driver:            d,
		},
	}
}

type PokyProvisioner struct {
	GenericProvisioner
}

func (provisioner *PokyProvisioner) String() string {
	return "Poky (Yocto Project Reference Distro)"
}

func (provisioner *PokyProvisioner) CompatibleWithHost() bool {
	const FirstYoctoVersion = 0.9
	isPoky := provisioner.OsReleaseInfo.ID == provisioner.OsReleaseID
	if !isPoky {
		return false
	}
	versionNumber, err := strconv.ParseFloat(provisioner.OsReleaseInfo.VersionID, 64)
	if err != nil {
		return false
	}

	return versionNumber >= FirstYoctoVersion

}

func (provisioner *PokyProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	command := fmt.Sprintf("sudo /etc/init.d/%s %s", name, action.String())

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *PokyProvisioner) Package(name string, action pkgaction.PackageAction) error {
	// Since no default package manager is installed, it does not support package function.
	return errors.New("poky provisioner does not support package")
}

func (provisioner *PokyProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *PokyProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	storageDriver, err := decideStorageDriver(provisioner, "aufs", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	// On Yocto projects, no default package manager is installed. To support Docker, developers need
	//  to install at build-time. So, here we do not install packages, and just check if Docker exist.
	log.Info("Check if Docker exist...")
	if err := mcnutils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
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
