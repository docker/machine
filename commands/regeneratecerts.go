package commands

import (
	"github.com/docker/machine/cli"
	"github.com/docker/machine/libmachine/log"
)

func cmdRegenerateCerts(c *cli.Context) {
	force := c.Bool("force")
	if force || confirmInput("Regenerate TLS machine certs?  Warning: this is irreversible.") {
		log.Infof("Regenerating TLS certificates")
		if err := runActionWithContext("configureAuth", c); err != nil {
			fatal(err)
		}
	}
}
