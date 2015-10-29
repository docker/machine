package commands

import "github.com/docker/machine/cli"

func cmdStop(c *cli.Context) error {
	return runActionWithContext("stop", c)
}
