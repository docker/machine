package commands

import (
	"github.com/docker/machine/libmachine/log"

	"github.com/codegangsta/cli"
)

func cmdStop(c *cli.Context) {
	if err := runActionWithContext("stop", c); err != nil {
		log.Fatal(err)
	}
}
