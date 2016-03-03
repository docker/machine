package provision

import (
	"fmt"
	"time"

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
	Register(drivers.WINDOWS, &RegisteredProvisioner{
		New: NewWindowsProvisioner,
	})
}

func NewWindowsProvisioner(d drivers.Driver) Provisioner {
	return &WindowsProvisioner{
		GenericProvisioner{
			Driver: d,
		},
	}
}

type WindowsProvisioner struct {
	GenericProvisioner
}

func (provisioner *WindowsProvisioner) String() string {
	return drivers.WINDOWS
}

func (provisioner *WindowsProvisioner) CompatibleWithHost() bool {
	if provisioner.Driver.GetOS() == drivers.WINDOWS {
		return true
	}
	return false
}

func (provisioner *WindowsProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	ip, err := provisioner.Driver.GetIP()
	if err != nil {
		return err
	}
	d := provisioner.Driver
	if out, err := drivers.WinRMRunCmd(ip, d.GetSSHUsername(), "docker#123$", fmt.Sprintf("net %s %s", action.String(), name)); err != nil {
		log.Debugf("service output", out)
		return err
	}
	return nil
}

func (provisioner *WindowsProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

func (provisioner *WindowsProvisioner) GetDockerOptionsDir() string {
	return "C:\\ProgramData\\docker\\certs.d"
}

func (provisioner *WindowsProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	ip, err := provisioner.Driver.GetIP()
	if err != nil {
		return false
	}

	d := provisioner.Driver
	dockerVersionCommand := "docker -H 127.0.0.1:2376 --tls --tlscert c:\\ProgramData\\docker\\certs.d\\server.pem --tlskey c:\\ProgramData\\docker\\certs.d\\server-key.pem --tlscacert c:\\ProgramData\\docker\\certs.d\\ca.pem version"

	if out, err := drivers.WinRMRunCmd(ip, d.GetSSHUsername(), "docker#123$", dockerVersionCommand); err != nil {
		log.Warnf("Error getting WinRM command to check if the daemon is up: %s", err)
		log.Debugf("docker version output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *WindowsProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.AuthOptions = authOptions

	log.Debug("Waiting for docker daemon")
	if err := mcnutils.WaitForSpecific(provisioner.dockerDaemonResponding, 60, 3*time.Second); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := ConfigureAuth(provisioner); err != nil {
		return err
	}

	return nil
}
