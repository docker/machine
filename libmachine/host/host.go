package host

import (
	"errors"
	"regexp"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/proxy"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
)

var (
	validHostNamePattern                               = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]*$`)
	errMachineMustBeRunningForUpgrade                  = errors.New("Error: machine must be running to upgrade.")
	stdSSHClientCreator               SSHClientCreator = &StandardSSHClientCreator{}
)

type SSHClientCreator interface {
	CreateSSHClient(d drivers.Driver) (ssh.Client, error)
}

type StandardSSHClientCreator struct {
	drivers.Driver
}

func SetSSHClientCreator(creator SSHClientCreator) {
	stdSSHClientCreator = creator
}

type Host struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *Options
	Name          string
	RawDriver     []byte `json:"-"`
}

type Options struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.Options
	SwarmOptions  *swarm.Options
	AuthOptions   *auth.Options
	SSHOptions    *ssh.Options
	ProxyOptions  *proxy.Options
}

type Metadata struct {
	ConfigVersion int
	DriverName    string
	HostOptions   Options
}

func ValidateHostName(name string) bool {
	return validHostNamePattern.MatchString(name)
}

func (h *Host) RunSSHCommand(command string) (string, error) {
	return drivers.RunSSHCommandFromDriver(h.Driver, command)
}

func (h *Host) CreateSSHClient() (ssh.Client, error) {
	return stdSSHClientCreator.CreateSSHClient(h.Driver)
}

func (creator *StandardSSHClientCreator) CreateSSHClient(d drivers.Driver) (ssh.Client, error) {
	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	options := &ssh.Options{}
	if d.GetSSHKeyPath() != "" {
		options.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewClient(d.GetSSHUsername(), addr, port, options)
}

func (h *Host) runActionForState(action func() error, desiredState state.State) error {
	if drivers.MachineInState(h.Driver, desiredState)() {
		return mcnerror.ErrHostAlreadyInState{
			Name:  h.Name,
			State: desiredState,
		}
	}

	if err := action(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(h.Driver, desiredState))
}

func (h *Host) WaitForDocker() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	return provision.WaitForDocker(provisioner, engine.DefaultPort)
}

func (h *Host) Start() error {
	log.Infof("Starting %q...", h.Name)
	if err := h.runActionForState(h.Driver.Start, state.Running); err != nil {
		return err
	}

	log.Infof("Machine %q was started.", h.Name)

	return h.WaitForDocker()
}

func (h *Host) Stop() error {
	log.Infof("Stopping %q...", h.Name)
	if err := h.runActionForState(h.Driver.Stop, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was stopped.", h.Name)
	return nil
}

func (h *Host) Kill() error {
	log.Infof("Killing %q...", h.Name)
	if err := h.runActionForState(h.Driver.Kill, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was killed.", h.Name)
	return nil
}

func (h *Host) Restart() error {
	log.Infof("Restarting %q...", h.Name)
	if drivers.MachineInState(h.Driver, state.Stopped)() {
		if err := h.Start(); err != nil {
			return err
		}
	} else if drivers.MachineInState(h.Driver, state.Running)() {
		if err := h.Driver.Restart(); err != nil {
			return err
		}
		if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
			return err
		}
	}

	return h.WaitForDocker()
}

func (h *Host) Upgrade() error {
	machineState, err := h.Driver.GetState()
	if err != nil {
		return err
	}

	if machineState != state.Running {
		return errMachineMustBeRunningForUpgrade
	}

	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	log.Info("Upgrading docker...")
	if err := provisioner.Package("docker", pkgaction.Upgrade); err != nil {
		return err
	}

	log.Info("Restarting docker...")
	return provisioner.Service("docker", serviceaction.Restart)
}

func (h *Host) URL() (string, error) {
	return h.Driver.GetURL()
}

func (h *Host) AuthOptions() *auth.Options {
	if h.HostOptions == nil {
		return nil
	}
	return h.HostOptions.AuthOptions
}

func (h *Host) ConfigureAuth() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	// TODO: This is kind of a hack (or is it?  I'm not really sure until
	// we have more clearly defined outlook on what the responsibilities
	// and modularity of the provisioners should be).
	//
	// Call provision to re-provision the certs properly.
	return provisioner.Provision(swarm.Options{}, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}

func (h *Host) Provision() error {
	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	return provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions)
}

func (h *Host) GetSocksProxy() string {
	if h.HostOptions != nil && h.HostOptions.ProxyOptions != nil {
		return h.HostOptions.ProxyOptions.SocksProxy
	}
	return ""
}
