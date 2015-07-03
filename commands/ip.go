package commands

import (
	"github.com/codegangsta/cli"
	"github.com/docker/machine/log"
)

func cmdIp(c *cli.Context) {
	if len(c.Args()) == 0 {
		cli.ShowCommandHelp(c, "ip")
		log.Fatal("You must specify a machine name")
	}

	ctx := "ip"

	if c.Bool("private") {
		ctx = "privateIp"
	}

	if err := runActionWithContext(ctx, c); err != nil {
		log.Fatal(err)
	}
}
