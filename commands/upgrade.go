package commands

import "github.com/docker/machine/cli"

func cmdUpgrade(c *cli.Context) {
	if err := runActionWithContext("upgrade", c); err != nil {
		fatal(err)
	}
}
