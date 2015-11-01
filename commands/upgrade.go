package commands

import "github.com/docker/machine/cli"

func cmdUpgrade(c *cli.Context) error {
	return runActionWithContext("upgrade", c)
}
