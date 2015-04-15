package commands

import (
	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdKill(c *cli.Context) {
	if err := runActionWithContext("kill", c); err != nil {
		log.Fatal(err)
	}
}
