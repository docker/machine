package provision

import (
	"github.com/docker/machine/drivers"
)

var provisioners = make(map[string]*ProvisionerDetection)

// Detection
type ProvisionerDetection struct {
	New ProvisionerFactoryFunc
}

func RegisterProvisioner(name string, p *ProvisionerDetection) {
	provisioners[name] = p
}

func DetectProvisioner(d drivers.Driver) (Provisioner, error) {
	for _, p := range provisioners {
		provisioner := p.New(d)

		if err := provisioner.CompatibleWithHost(); err != nil {
			return provisioner, nil
		}
	}

	return nil, ErrDetectionFailed
}

type ProvisionerFactoryFunc func(drivers.Driver) Provisioner

// Distribution specific actions
type Provisioner interface {
	// Perform action on a named service
	Service(name string, action ServiceState) error

	// Ensure a package state
	Package(name string, action PackageState) error

	// Hostname
	Hostname() string

	// Set hostname
	SetHostname(hostname string) error

	// Detection function
	CompatibleWithHost() error
}
