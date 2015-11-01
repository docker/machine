package provision

import (
	"github.com/docker/machine/libmachine/drivers"
)

func init() {
	Register("Centos", &RegisteredProvisioner{
		New: NewCentosProvisioner,
	})
}

func NewCentosProvisioner(d drivers.Driver) Provisioner {
	g := GenericProvisioner{
		DockerOptionsDir:  "/etc/docker",
		DaemonOptionsFile: "/etc/systemd/system/docker.service",
		OsReleaseId:       "centos",
		Packages:          []string{},
		Driver:            d,
	}
	p := &CentosProvisioner{
		RedHatProvisioner{
			GenericProvisioner: g,
		},
	}
	return p
}

type CentosProvisioner struct {
	RedHatProvisioner
}
