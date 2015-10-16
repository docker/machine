package commands

import "github.com/docker/machine/cli"

func cmdKill(c *cli.Context) {
	if err := runActionWithContext("kill", c); err != nil {
		fatal(err)
	}
}
