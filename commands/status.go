package commands

import (
	"github.com/docker/machine/cli"
	"github.com/docker/machine/libmachine/log"
)

func cmdStatus(c *cli.Context) {
	if len(c.Args()) != 1 {
		fatal(ErrExpectedOneMachine)
	}

	host := getFirstArgHost(c)
	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Errorf("error getting state for host %s: %s", host.Name, err)
	}

	log.Info(currentState)
}
