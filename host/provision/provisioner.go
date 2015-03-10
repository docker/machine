package provision

import (
	"os/exec"

	"github.com/docker/machine/drivers"
)

var provisioners = make(map[string]*ProvisionerFactories)

// Detection
type ProvisionerFactories struct {
	New ProvisionerFactoryFunc
}

func RegisterProvisioner(name string, p *ProvisionerFactories) {
	provisioners[name] = p
}

func DetectProvisioner(d drivers.Driver) (Provisioner, error) {
	for _, p := range provisioners {
		provisioner := p.New(d.GetSSHCommand)

		if err := provisioner.CompatibleWithHost(); err == nil {
			return provisioner, nil
		}
	}

	return nil, ErrDetectionFailed
}

type SSHCommandFunc func(args ...string) (*exec.Cmd, error)
type ProvisionerFactoryFunc func(SSHCommandFunc) Provisioner

// Distribution specific actions
type Provisioner interface {
	// Perform action on a named service
	Service(name string, action ServiceState) error

	// Ensure a package state
	Package(name string, action PackageState) error

	// Hostname
	Hostname() (string, error)

	// Set hostname
	SetHostname(hostname string) error

	// Detection function
	CompatibleWithHost() error
}
