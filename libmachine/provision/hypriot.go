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
	Register("Hypriot", &RegisteredProvisioner{
		New: NewHypriotProvisioner,
	})
}

func NewHypriotProvisioner(d drivers.Driver) Provisioner {
	return &HypriotProvisioner{
		GenericProvisioner{
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/default/docker",
			OsReleaseId:       "raspbian",
			Packages:          []string{},
			Driver:            d,
		},
	}
}

type HypriotProvisioner struct {
	GenericProvisioner
}

func (provisioner *HypriotProvisioner) Service(name string, action pkgaction.ServiceAction) error {
	command := fmt.Sprintf("sudo service %s %s", name, action.String())

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *HypriotProvisioner) Package(name string, action pkgaction.PackageAction) error {
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
		packageAction = "install --only-upgrade"
	}

	switch name {
	case "docker":
		name = "docker-hypriot"
	}

	if updateMetadata {
		// invoke apt-get update for metadata
		if _, err := provisioner.SSHCommand("sudo -E apt-get update"); err != nil {
			return err
		}
	}

	command := fmt.Sprintf("DEBIAN_FRONTEND=noninteractive sudo -E apt-get %s -y %s", packageAction, name)

	if _, err := provisioner.SSHCommand(command); err != nil {
		return err
	}

	return nil
}

func (provisioner *HypriotProvisioner) dockerDaemonResponding() bool {
	if _, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is up: %s", err)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *HypriotProvisioner) dockerDaemonInstalled() bool {
	if _, err := provisioner.SSHCommand("type docker"); err != nil {
		log.Warnf("Docker not installed, let's install it")
		return false
	}

	return true
}

func (provisioner *HypriotProvisioner) dockerDaemonRunning() bool {
	if _, err := provisioner.SSHCommand("sudo service docker status"); err != nil {
		log.Warnf("Docker not running")
		return false
	}

	return true
}

func (provisioner *HypriotProvisioner) setHostnameHypriot(hostname string) error {
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"if [ -f /boot/occidentalis.txt ]; then sudo sed -i 's/^hostname.*=.*/hostname=%s/g' /boot/occidentalis.txt; fi",
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *HypriotProvisioner) setHypriotAptRepo() error {
	if _, err := provisioner.SSHCommand("if [ ! -f /etc/apt/sources.list.d/hypriot.list ]; then echo 'deb http://repository.hypriot.com/ wheezy main' | sudo tee /etc/apt/sources.list.d/hypriot.list; fi"); err != nil {
		return err
	}

	return nil
}

func (provisioner *HypriotProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions, engineOptions engine.EngineOptions) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions

	if provisioner.EngineOptions.StorageDriver == "" {
		provisioner.EngineOptions.StorageDriver = "overlay"
	}

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if err := provisioner.setHostnameHypriot(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if err := provisioner.setHypriotAptRepo(); err != nil {
		return err
	}

	if !provisioner.dockerDaemonInstalled() {
		provisioner.Packages = append(provisioner.Packages, "docker")
	}

	for _, pkg := range provisioner.Packages {
		if err := provisioner.Package(pkg, pkgaction.Install); err != nil {
			return err
		}
	}

	if !provisioner.dockerDaemonRunning() {
		if err := provisioner.Service("docker", pkgaction.Start); err != nil {
			return err
		}
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
