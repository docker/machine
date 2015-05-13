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
		DebianFamilyProvisioner{
			GenericProvisioner{
				OsReleaseId:       "ubuntu",
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

type UbuntuProvisioner struct {
	DebianFamilyProvisioner
}
