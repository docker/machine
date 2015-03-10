package provisioner

import (
	"fmt"

	"github.com/docker/machine/drivers"
)

func init() {
	RegisterRuntime("boot2docker", &RegisteredRuntime{
		Detect: Boot2DockerDetection,
	})
}

func Boot2DockerDetection(d drivers.Driver) (Runtime, error) {
	_, err := sshCommand(d, "lsb_release")
	if err != nil {
		fmt.Print(err)
		return nil, ErrDetectionFailed
	}

	return Boot2Docker{}, nil
}

type Boot2Docker struct{}

func (r Boot2Docker) Service(name string, state ServiceState) error {
	return nil
}

func (r Boot2Docker) Package(name string, state PackageState) error {
	return ErrNotImplemented
}
