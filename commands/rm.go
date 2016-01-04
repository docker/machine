package commands

import (
	"fmt"

	"strings"

	"errors"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func cmdRm(c CommandLine, api libmachine.API) error {
	if len(c.Args()) == 0 {
		c.ShowHelp()
		return ErrNoMachineSpecified
	}

	log.Info(fmt.Sprintf("About to remove %s", strings.Join(c.Args(), ",")))

	force := c.Bool("force")
	confirm := c.Bool("y")
	var errorOccured string

	if !userConfirm(confirm, force) {
		return nil
	}

	for _, hostName := range c.Args() {
		err := removeRemoteMachine(hostName, api)
		if err != nil {
			errorOccured = fmt.Sprintf("Error removing host %q: %s", hostName, err)
			log.Errorf(errorOccured)
		}

		if err == nil || force {
			removeErr := api.Remove(hostName)
			if removeErr != nil {
				log.Errorf("Error removing machine %q from store: %s", hostName, removeErr)
			} else {
				log.Infof("Successfully removed %s", hostName)
			}
		}
	}
	if errorOccured != "" {
		return errors.New(errorOccured)
	}
	return nil
}

func userConfirm(confirm bool, force bool) bool {
	if confirm || force {
		return true
	}

	sure, err := confirmInput(fmt.Sprintf("Are you sure?"))
	if err != nil {
		return false
	}

	return sure
}

func removeRemoteMachine(hostName string, api libmachine.API) error {
	currentHost, loaderr := api.Load(hostName)
	if loaderr != nil {
		return loaderr
	}

	return currentHost.Driver.Remove()
}
