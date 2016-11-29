package commands

import (
	"github.com/docker/machine/libmachine"
)

func cmdReinstall(c CommandLine, api libmachine.API) error {
	if !c.Bool("force") {
		ok, err := confirmInput("Reinstall remote instances?  Warning: this is irreversible.")
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
	}

	return runAction("reinstall", c, api)
}
