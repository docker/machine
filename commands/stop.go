package commands

import "github.com/docker/machine/cli"

func cmdStop(c *cli.Context) {
	if err := runActionWithContext("stop", c); err != nil {
		fatal(err)
	}
}
