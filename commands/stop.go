package commands

import (
	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
)

func cmdStop(c *cli.Context) {
	log.Spinner.Start("Stopping machine(s)")
	if err := runActionWithContext("stop", c); err != nil {
		log.Fatal(err)
	}
	log.Spinner.Stop()
}
