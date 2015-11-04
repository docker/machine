package commands

import "github.com/docker/machine/cli"

func cmdIP(c *cli.Context) error {
	return runActionWithContext("ip", c)
}
