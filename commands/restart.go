package commands

import (
	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
)

func cmdRestart(c *cli.Context) {
	log.Spinner.Start("Restarting machine(s)")
	if err := runActionWithContext("restart", c); err != nil {
		log.Fatal(err)
	}
	log.Spinner.Stop()
	log.Info("Restarted machines may have new IP addresses. You may need to re-run the `docker-machine env` command.")
}
