package commands

import (
	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
)

func cmdStatus(c *cli.Context) {
	host := getHost(c)
	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Fatal(err)
	}
	log.Info(currentState)
}
