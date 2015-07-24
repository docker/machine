package provision

import (
	"errors"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/log"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

var (
	ErrUnknownDriver = errors.New("unknown driver")
)

func init() {
	Register("boot2docker", &RegisteredProvisioner{
		New: NewBoot2DockerProvisioner,
	})
}

func NewBoot2DockerProvisioner(d drivers.Driver) Provisioner {
	g := GenericProvisioner{
		DockerOptionsDir:  "/etc/docker",
		DaemonOptionsFile: "/etc/systemd/system/docker.service",
		OsReleaseId:       "docker",
		Packages:          []string{},
		Driver:            d,
	}
	p := &Boot2DockerProvisioner{
		DebianProvisioner{
			GenericProvisioner: g,
		},
	}
	return p
}

type Boot2DockerProvisioner struct {
	DebianProvisioner
}

func (provisioner *Boot2DockerProvisioner) upgradeIso() error {
	log.Info("Stopping machine to do the upgrade...")

	if err := provisioner.Driver.Stop(); err != nil {
		return err

	}

	if err := utils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Stopped)); err != nil {
		return err

	}

	machineName := provisioner.GetDriver().GetMachineName()

	log.Infof("Upgrading machine %s...", machineName)

	isoFilename := ""
	switch provisioner.GetDriver().DriverName() {
	case "virtualbox":
		isoFilename = "boot2docker-virtualbox.iso"
	case "vmwarefusion", "vmwarevsphere":
		isoFilename = "boot2docker-vmware.iso"
	case "hyper-v":
		isoFilename = "boot2docker-hyperv.iso"
	default:
		return ErrUnknownDriver
	}

	b2dutils := utils.NewB2dUtils("", "", isoFilename)

	// Usually we call this implicitly, but call it here explicitly to get
	// the latest boot2docker ISO.
	if err := b2dutils.DownloadLatestBoot2Docker(); err != nil {
		return err

	}

	// Copy the latest version of boot2docker ISO to the machine's directory
	if err := b2dutils.CopyIsoToMachineDir("", machineName); err != nil {
		return err

	}

	if err := provisioner.Driver.Start(); err != nil {
		return err

	}

	return utils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Running))

}

func (provisioner *Boot2DockerProvisioner) Package(name string, action pkgaction.PackageAction) error {
	if name == "docker" && action == pkgaction.Upgrade {
		if err := provisioner.upgradeIso(); err != nil {
			return err

		}

	}
	return nil

}
