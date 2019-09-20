package provision

import (
	"github.com/docker/machine/libmachine/drivers"
)

func init() {
	Register("Amazon2", &RegisteredProvisioner{
		New: NewAmazon2Provisioner,
	})
}

func NewAmazon2Provisioner(d drivers.Driver) Provisioner {
	return &Amazon2Provisioner{
		NewRedHatProvisioner("amazon2", d),
	}
}

type Amazon2Provisioner struct {
	*RedHatProvisioner
}

func (provisioner *Amazon2Provisioner) String() string {
	return "amazon2"
}
