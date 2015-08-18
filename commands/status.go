package commands

import (
	"github.com/docker/machine/libmachine/log"

	"github.com/codegangsta/cli"
)

func cmdStatus(c *cli.Context) {
	host := getFirstArgHost(c)
	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Errorf("error getting state for host %s: %s", host.Name, err)
	}
	log.Info(currentState)
}
