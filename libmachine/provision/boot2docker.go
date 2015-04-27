package provision

import (
	"bytes"
	"errors"
	"fmt"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

func init() {
	Register("boot2docker", &RegisteredProvisioner{
		New: NewBoot2DockerProvisioner,
	})
}

func NewBoot2DockerProvisioner(d drivers.Driver) Provisioner {
	return &Boot2DockerProvisioner{
		Driver: d,
	}
}

type Boot2DockerProvisioner struct {
	OsReleaseInfo *OsRelease
	Driver        drivers.Driver
	SwarmOptions  swarm.SwarmOptions
}

func (provisioner *Boot2DockerProvisioner) Service(name string, action pkgaction.ServiceAction) error {
	var (
		err error
	)

	if _, err = provisioner.SSHCommand(fmt.Sprintf("sudo /etc/init.d/%s %s", name, action.String())); err != nil {
		return err
	}

	return nil
}

func (provisioner *Boot2DockerProvisioner) upgradeIso() error {
	switch provisioner.Driver.DriverName() {
	case "vmwarefusion", "vmwarevsphere":
		return errors.New("Upgrade functionality is currently not supported for these providers, as they use a custom ISO.")
	}

	log.Info("Stopping machine to do the upgrade...")

	if err := provisioner.Driver.Stop(); err != nil {
		return err
	}

	if err := utils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Stopped)); err != nil {
		return err
	}

	machineName := provisioner.GetDriver().GetMachineName()

	log.Infof("Upgrading machine %s...", machineName)

	b2dutils := utils.NewB2dUtils("", "")

	// Usually we call this implicitly, but call it here explicitly to get
	// the latest boot2docker ISO.
	if err := b2dutils.DownloadLatestBoot2Docker(); err != nil {
		return err
	}

	// Copy the latest version of boot2docker ISO to the machine's directory
	if err := b2dutils.CopyIsoToMachineDir("", machineName); err != nil {
		return err
	}

	log.Infof("Starting machine back up...")

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

func (provisioner *Boot2DockerProvisioner) Hostname() (string, error) {
	output, err := provisioner.SSHCommand(fmt.Sprintf("hostname"))
	if err != nil {
		return "", err
	}

	var so bytes.Buffer
	if _, err := so.ReadFrom(output.Stdout); err != nil {
		return "", err
	}

	return so.String(), nil
}

func (provisioner *Boot2DockerProvisioner) SetHostname(hostname string) error {
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /var/lib/boot2docker/etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *Boot2DockerProvisioner) GetDockerOptionsDir() string {
	return "/var/lib/boot2docker"
}

func (provisioner *Boot2DockerProvisioner) GenerateDockerOptions(dockerPort int, authOptions auth.AuthOptions) (*DockerOptions, error) {
	defaultDaemonOpts := getDefaultDaemonOpts(provisioner.Driver.DriverName(), authOptions)
	daemonOpts := fmt.Sprintf("-H tcp://0.0.0.0:%d", dockerPort)
	daemonOptsDir := path.Join(provisioner.GetDockerOptionsDir(), "profile")
	opts := fmt.Sprintf("%s %s", defaultDaemonOpts, daemonOpts)
	daemonCfg := fmt.Sprintf(`EXTRA_ARGS='%s'
CACERT=%s
SERVERCERT=%s
SERVERKEY=%s
DOCKER_TLS=no`, opts, authOptions.CaCertRemotePath, authOptions.ServerCertRemotePath, authOptions.ServerKeyRemotePath)
	return &DockerOptions{
		EngineOptions:     daemonCfg,
		EngineOptionsPath: daemonOptsDir,
	}, nil
}

func (provisioner *Boot2DockerProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.Id == "boot2docker"
}

func (provisioner *Boot2DockerProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *Boot2DockerProvisioner) Provision(swarmOptions swarm.SwarmOptions, authOptions auth.AuthOptions) error {
	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if err := installDockerGeneric(provisioner); err != nil {
		return err
	}

	ip, err := provisioner.GetDriver().GetIP()
	if err != nil {
		return err
	}

	// b2d hosts need to wait for the daemon to be up
	// before continuing with provisioning
	if err := utils.WaitForDocker(ip, 2376); err != nil {
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

func (provisioner *Boot2DockerProvisioner) SSHCommand(args string) (ssh.Output, error) {
	return drivers.RunSSHCommandFromDriver(provisioner.Driver, args)
}

func (provisioner *Boot2DockerProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
