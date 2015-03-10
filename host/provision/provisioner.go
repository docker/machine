package provision

import (
	"fmt"
	"github.com/docker/machine/drivers"
)

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
}

// var runtimes = make(map[string]*RegisteredRuntime)

// type RegisteredRuntime struct {
// 	Detect DetectionFunc
// }

// func RegisterRuntime(name string, runtime *RegisteredRuntime) error {
// 	runtimes[name] = runtime

// 	return nil
// }

// type DetectionFunc func(d drivers.Driver) (Runtime, error)

// func DetectRuntime(d drivers.Driver) (Runtime, error) {
// 	for _, r := range runtimes {
// 		return r.Detect(d)
// 	}

// 	return nil, ErrDetectionFailed
// }
