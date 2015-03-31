package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func cmdRegenerateCerts(c *cli.Context) {
	force := c.Bool("force")
	if force || confirmInput("Regenerate TLS machine certs?  Warning: this is irreversible.") {
		log.Infof("Regenerating TLS certificates")
		if err := runActionWithContext("configureAuth", c); err != nil {
			log.Fatal(err)
		}
	}
}
