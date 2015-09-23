package persist

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
)

type Store interface {
	// Exists returns whether a machine exists or not
	Exists(name string) (bool, error)

	// NewHost will initialize a new host machine
	NewHost(driver drivers.Driver) (*host.Host, error)

	// List returns a list of all hosts in the store
	List() ([]*host.Host, error)

	// Get loads a host by name
	Load(name string) (*host.Host, error)

	// Remove removes a machine from the store
	Remove(name string, force bool) error

	// Save persists a machine in the store
	Save(host *host.Host) error
}
