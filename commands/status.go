package commands

import (
	"fmt"

	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
)

func cmdStatus(c *cli.Context) {
	host := getHost(c)
	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Errorf("error getting state for host %s: %s", host.Name, err)
	}
	fmt.Println(currentState)
}
