package commands

import "github.com/docker/machine/cli"

func cmdIp(c *cli.Context) {
	if err := runActionWithContext("ip", c); err != nil {
		fatal(err)
	}
}
