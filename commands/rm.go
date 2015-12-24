package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func cmdRm(c CommandLine, api libmachine.API) error {
	if len(c.Args()) == 0 {
		c.ShowHelp()
		return ErrNoMachineSpecified
	}

	force := c.Bool("force")
	confirm := c.Bool("y")

	for _, hostName := range c.Args() {
		h, loaderr := api.Load(hostName)
		if loaderr != nil {
			// On --force, continue to remove on-disk files/dir
			if !force {
				return fmt.Errorf("Error removing host %q: %s", hostName, loaderr)
			}
			log.Errorf("Error removing host %q: %s. Continuing on `-f`, host instance may by running", hostName, loaderr)
		}
		defer api.Close(h)

		if !confirm && !force {
			userinput, err := confirmInput(fmt.Sprintf("Do you really want to remove %q?", hostName))
			if !userinput {
				return err
			}
		}

		if loaderr == nil {
			if err := h.Driver.Remove(); err != nil {
				if !force {
					log.Errorf("Provider error removing machine %q: %s", hostName, err)
					continue
				}
			}
		}

		if err := api.Remove(hostName); err != nil {
			log.Errorf("Error removing machine %q from store: %s", hostName, err)
		} else {
			log.Infof("Successfully removed %s", hostName)
		}
	}

	return nil
}
