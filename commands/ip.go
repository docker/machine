package commands

import "github.com/docker/machine/cli"

func cmdIp(c *cli.Context) error {
	return runActionWithContext("ip", c)
}
