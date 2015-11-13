package commands

import (
	"errors"
	"fmt"

	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/log"
)

func cmdRm(cli CommandLine, store rpcdriver.Store) error {
	if len(cli.Args()) == 0 {
		cli.ShowHelp()
		return errors.New("You must specify a machine name")
	}

	force := cli.Bool("force")

	for _, hostName := range cli.Args() {
		h, err := store.Load(hostName)
		if err != nil {
			return fmt.Errorf("Error removing host %q: %s", hostName, err)
		}

		if err := h.Driver.Remove(); err != nil {
			if !force {
				log.Errorf("Provider error removing machine %q: %s", hostName, err)
				continue
			}
		}

		if err := store.Remove(hostName); err != nil {
			log.Errorf("Error removing machine %q from store: %s", hostName, err)
		} else {
			log.Infof("Successfully removed %s", hostName)
		}
	}

	return nil
}
