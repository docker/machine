package commands

import (
	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdIp(c *cli.Context) {
	if err := runActionWithContext("ip", c); err != nil {
		log.Fatal(err)
	}
}
