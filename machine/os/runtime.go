package os

import (
	"github.com/docker/machine/drivers"
)

type ServiceState int

const (
	Restart ServiceState = iota
	Start
	Stop
)

var serviceStates = []string{
	"restart",
	"start",
	"stop",
}

func (s ServiceState) String() string {
	if int(s) >= 0 && int(s) < len(serviceStates) {
		return serviceStates[s]
	}

	return ""
}

type PackageState int

const (
	Installed PackageState = iota
	Missing
)

var packageStates = []string{
	"installed",
	"missing",
}

func (s PackageState) String() string {
	if int(s) >= 0 && int(s) < len(packageStates) {
		return packageStates[s]
	}

	return ""
}

// Distribution specific actions
type Runtime interface {
	// Perform action on a named service
	Service(name string, action ServiceState) error

	// Ensure a package state
	Package(name string, action PackageState) error
}

var (
	runtimes map[string]*RegisteredRuntime
)

func init() {
	runtimes = make(map[string]*RegisteredRuntime)
}

type RegisteredRuntime struct {
	Detect DetectionFunc
}

func RegisterRuntime(name string, runtime *RegisteredRuntime) error {
	return nil
}

type DetectionFunc func(d drivers.Driver) (*Runtime, error)

func sshCommand(d drivers.Driver, args ...string) (string, error) {
	cmd, err := d.GetSSHCommand(args...)
	if err != nil {
		return "", ErrSSHCommandFailed
	}

	if err := cmd.Run(); err != nil {
		return "", ErrSSHCommandFailed
	}

	out, err := cmd.Output()
	if err != nil {
		return "", ErrSSHCommandFailed
	}

	return string(out), nil
}
