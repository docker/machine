package commands

import "github.com/docker/machine/cli"

func cmdIP(c *cli.Context) {
	if err := runActionWithContext("ip", c); err != nil {
		fatal(err)
	}
}
