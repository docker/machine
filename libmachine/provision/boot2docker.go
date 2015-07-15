package provision

import (
	"github.com/docker/machine/drivers"
)

func init() {
	Register("boot2docker", &RegisteredProvisioner{
		New: NewBoot2DockerProvisioner,
	})
}

func NewBoot2DockerProvisioner(d drivers.Driver) Provisioner {
	g := GenericProvisioner{
		DockerOptionsDir:  "/etc/docker",
		DaemonOptionsFile: "/etc/systemd/system/docker.service",
		OsReleaseId:       "docker",
		Packages:          []string{},
		Driver:            d,
	}
	p := &Boot2DockerProvisioner{
		DebianProvisioner{
			GenericProvisioner: g,
		},
	}
	return p
}

type Boot2DockerProvisioner struct {
	DebianProvisioner
}
