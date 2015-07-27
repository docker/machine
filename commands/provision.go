package commands

import (
	"log"

	"github.com/codegangsta/cli"
)

func cmdProvision(c *cli.Context) {
	if err := runActionWithContext("provision", c); err != nil {
		log.Fatal(err)
	}
}
