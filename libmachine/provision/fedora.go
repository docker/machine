package provision

import (
	"github.com/docker/machine/drivers"
)

const (
	// TODO: eventually the RPM install process will be integrated
	// into the get.docker.com install script; for now
	// we install via vendored RPMs
	dockerFedoraRPMPath = "https://docker-mcn.s3.amazonaws.com/public/fedora/rpms/docker-engine-1.6.1-0.0.20150511.171646.git1b47f9f.fc21.x86_64.rpm"
)

func init() {
	Register("Fedora", &RegisteredProvisioner{
		New: NewFedoraProvisioner,
	})
}

func NewFedoraProvisioner(d drivers.Driver) Provisioner {
	g := GenericProvisioner{
		DockerOptionsDir:  "/etc/docker",
		DaemonOptionsFile: "/lib/systemd/system/docker.service",
		OsReleaseId:       "fedora",
		Packages: []string{
			"curl",
			"yum-utils",
		},
		Driver: d,
	}
	p := &FedoraProvisioner{
		RedHatProvisioner{
			g,
			dockerFedoraRPMPath,
		},
	}
	return p
}

type FedoraProvisioner struct {
	RedHatProvisioner
}
