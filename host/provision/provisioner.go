package provision

import (
	"bytes"
	"fmt"
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

func DetectProvisioner(d drivers.Driver) (*Provisioner, error) {
	var (
		osReleaseOut bytes.Buffer
	)
	catOsReleaseCmd, err := d.GetSSHCommand("cat /etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("Error getting SSH command: %s", err)
	}

	// Normally I would just use Output() for this, but d.GetSSHCommand
	// defaults to sending the output of the command to stdout in debug
	// mode, so that will be broken if we don't set it ourselves.
	catOsReleaseCmd.Stdout = &osReleaseOut

	if err := catOsReleaseCmd.Run(); err != nil {
		return nil, fmt.Errorf("Error running SSH command to get /etc/os-release: %s", err)
	}

	osReleaseInfo, err := NewOsRelease(osReleaseOut.Bytes())
	if err != nil {
		return nil, fmt.Errorf("Error parsing /etc/os-release file: %s", err)
	}

	for _, p := range provisioners {
		provisioner := p.New(d.GetSSHCommand)
		provisioner.SetOsReleaseInfo(osReleaseInfo)

		if provisioner.CompatibleWithHost() {
			return &provisioner, nil
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
	CompatibleWithHost() bool

	// Set the OS Release info depending on how it's represented
	// internally
	SetOsReleaseInfo(info *OsRelease)
}
