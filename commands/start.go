package commands

import (
	"github.com/docker/machine/libmachine/log"

	"github.com/docker/machine/cli"
)

func cmdStart(c *cli.Context) {
	if err := runActionWithContext("start", c); err != nil {
		fatal(err)
	}
	log.Info("Started machines may have new IP addresses. You may need to re-run the `docker-machine env` command.")
}
