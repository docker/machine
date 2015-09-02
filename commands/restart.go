package commands

import (
	"github.com/docker/machine/libmachine/log"

	"github.com/docker/machine/cli"
)

func cmdRestart(c *cli.Context) {
	if err := runActionWithContext("restart", c); err != nil {
		fatal(err)
	}
	log.Info("Restarted machines may have new IP addresses. You may need to re-run the `docker-machine env` command.")
}
