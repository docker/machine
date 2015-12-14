package commands

import (
	"fmt"
	"path/filepath"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/check"
	"github.com/docker/machine/libmachine/log"
)

const dindCertPath = "/etc/docker/cert"

func cmdConfig(c CommandLine, api libmachine.API) error {
	// Ensure that log messages always go to stderr when this command is
	// being run (it is intended to be run in a subshell)
	log.RedirectStdOutToStdErr()

	if len(c.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := api.Load(c.Args().First())
	if err != nil {
		return err
	}

	dockerHost, authOptions, err := check.DefaultConnChecker.Check(host, c.Bool("swarm"))
	if err != nil {
		return fmt.Errorf("Error running connection boilerplate: %s", err)
	}

	log.Debug(dockerHost)

	dind := c.Bool("dind")
	if dind {
		machineCertPath := filepath.Join(mcndirs.GetMachineDir(), host.Name)
		fmt.Printf("--volume=\"%s:%s:ro\"\n--env=\"DOCKER_CERT_PATH=%s\"\n--env=\"DOCKER_TLS_VERIFY=1\"\n--env=\"DOCKER_HOST=%s\"\n",
			machineCertPath, dindCertPath, dindCertPath, dockerHost)
	} else {
		fmt.Printf("--tlsverify\n--tlscacert=%q\n--tlscert=%q\n--tlskey=%q\n-H=%s\n",
			authOptions.CaCertPath, authOptions.ClientCertPath, authOptions.ClientKeyPath, dockerHost)
	}

	return nil
}
