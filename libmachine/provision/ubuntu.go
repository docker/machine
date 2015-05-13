package provision

import (
	"github.com/docker/machine/drivers"
)

func init() {
	Register("Ubuntu", &RegisteredProvisioner{
		New: NewUbuntuProvisioner,
	})
}

func NewUbuntuProvisioner(d drivers.Driver) Provisioner {
	return &UbuntuProvisioner{
		newDebianFamilyProvisioner(d, "ubuntu"),
	}
}

type UbuntuProvisioner struct {
	DebianFamilyProvisioner
}
