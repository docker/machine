package commands

import (
	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
)

func cmdStart(c *cli.Context) {
	log.Spinner.Start("Starting machine(s)")
	if err := runActionWithContext("start", c); err != nil {
		log.Fatal(err)
	}
	log.Spinner.Stop()
	log.Info("Started machines may have new IP addresses. You may need to re-run the `docker-machine env` command.")
}
