package provision

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision/serviceaction"
)

type UpstartProvisioner struct {
	GenericProvisioner
}

func (p *UpstartProvisioner) String() string {
	return "upstart"
}

func NewUpstartProvisioner(osReleaseID string, d drivers.Driver) UpstartProvisioner {
	return UpstartProvisioner{
		GenericProvisioner{
			SSHCommander:      GenericSSHCommander{Driver: d},
			DockerOptionsDir:  "/etc/docker",
			DaemonOptionsFile: "/etc/default/docker",
			OsReleaseID:       osReleaseID,
			Packages: []string{
				"curl",
			},
			Driver: d,
		},
	}
}

func (p *UpstartProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	command := fmt.Sprintf("sudo service %s %s", name, action.String())

	if _, err := p.SSHCommand(command); err != nil {
		return err
	}

	return nil
}
