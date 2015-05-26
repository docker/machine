package provision

import (
	"github.com/docker/machine/drivers"
)

func init() {
	Register("Debian", &RegisteredProvisioner{
		New: NewDebianProvisioner,
	})
}

func NewDebianProvisioner(d drivers.Driver) Provisioner {
	return &DebianProvisioner{
		newDebianFamilyProvisioner(d, "debian"),
	}
}

type DebianProvisioner struct {
	DebianFamilyProvisioner
}
