package commands

import (
	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdStart(c *cli.Context) {
	if err := runActionWithContext("start", c); err != nil {
		log.Fatal(err)
	}
}
