package commands

import (
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

func cmdStatus(cli CommandLine, store persist.Store) error {
	if len(cli.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := loadHost(store, cli.Args().First())
	if err != nil {
		return err
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Errorf("error getting state for host %s: %s", host.Name, err)
	}

	log.Info(currentState)

	return nil
}
