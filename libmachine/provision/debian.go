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
		DebianFamilyProvisioner{
			GenericProvisioner{
				OsReleaseId:       "debian",
				DockerOptionsDir:  "/etc/docker",
				DaemonOptionsFile: "/etc/default/docker",
				Packages: []string{
					"curl",
				},
				Driver: d,
			},
		},
	}
}

type DebianProvisioner struct {
	DebianFamilyProvisioner
}
