package commands

import (
	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
)

func cmdUpgrade(c *cli.Context) {
	if err := runActionWithContext("upgrade", c); err != nil {
		log.Fatal(err)
	}
}
