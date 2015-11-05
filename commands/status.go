package commands

import (
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

func cmdStatus(c CommandLine, store persist.Store) error {
	if len(c.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := store.Load(c.Args().First())
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
