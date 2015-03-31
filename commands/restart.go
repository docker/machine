package commands

import (
	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdRestart(c *cli.Context) {
	if err := runActionWithContext("restart", c); err != nil {
		log.Fatal(err)
	}
}
