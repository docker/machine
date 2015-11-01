package provision

import (
	"github.com/docker/machine/libmachine/drivers"
)

func init() {
	Register("Fedora", &RegisteredProvisioner{
		New: NewFedoraProvisioner,
	})
}

func NewFedoraProvisioner(d drivers.Driver) Provisioner {
	g := GenericProvisioner{
		DockerOptionsDir:  "/etc/docker",
		DaemonOptionsFile: "/etc/systemd/system/docker.service",
		OsReleaseId:       "fedora",
		Packages:          []string{},
		Driver:            d,
	}
	p := &FedoraProvisioner{
		RedHatProvisioner{
			GenericProvisioner: g,
		},
	}
	return p
}

type FedoraProvisioner struct {
	RedHatProvisioner
}
