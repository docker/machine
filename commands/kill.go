package commands

import "github.com/docker/machine/cli"

func cmdKill(c *cli.Context) error {
	return runActionWithContext("kill", c)
}
