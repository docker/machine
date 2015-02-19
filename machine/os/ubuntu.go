package os

import (
	"fmt"
	"strings"

	"github.com/docker/machine/drivers"
)

func init() {
	RegisterRuntime("ubuntu", &RegisteredRuntime{
		Detect: UbuntuDetection,
	})
}

func UbuntuDetection(d drivers.Driver) (Runtime, error) {
	out, err := sshCommand(d, "lsb_release -i | cut -f2 -d:")
	if err != nil {
		return nil, ErrDetectionFailed
	}

	fmt.Println(out, out == "Ubuntu")

	out = strings.Trim(out, " ")
	if out != "Ubuntu" {
		return nil, ErrDetectionFailed
	}

	fmt.Println(out, out == "Ubuntu")

	return Ubuntu{}, nil
}

type Ubuntu struct{}

func (r Ubuntu) Service(name string, state ServiceState) error {
	return nil
}

func (r Ubuntu) Package(name string, state PackageState) error {
	return nil
}
