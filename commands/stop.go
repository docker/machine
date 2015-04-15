package commands

import (
	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdStop(c *cli.Context) {
	if err := runActionWithContext("stop", c); err != nil {
		log.Fatal(err)
	}
}
