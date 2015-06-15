package provision

import (
	"github.com/docker/machine/drivers"
)

const (
	// TODO: eventually the RPM install process will be integrated
	// into the get.docker.com install script; for now
	// we install via vendored RPMs
	dockerFedoraRPMPath = "https://test.docker.com/rpm/1.7.0-rc3/fedora-21/RPMS/x86_64/docker-engine-1.7.0-0.3.rc3.fc21.x86_64.rpm"
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
			DockerRPMPath:      dockerFedoraRPMPath,
		},
	}
	return p
}

type FedoraProvisioner struct {
	RedHatProvisioner
}
