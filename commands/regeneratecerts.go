package commands

import (
	"github.com/docker/machine/libmachine/log"
)

func cmdRegenerateCerts(c CommandLine) error {
	if !c.Bool("force") {
		ok, err := confirmInput("Regenerate TLS machine certs?  Warning: this is irreversible.")
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
	}

	log.Infof("Regenerating TLS certificates")

	return runActionWithContext("configureAuth", c)
}
