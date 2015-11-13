package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

func cmdRegenerateCerts(cli CommandLine, store persist.Store) error {
	if !cli.Bool("force") {
		ok, err := confirmInput("Regenerate TLS machine certs?  Warning: this is irreversible.")
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
	}

	log.Infof("Regenerating TLS certificates")

	return runActionOnHosts(func(h *host.Host) error {
		return h.ConfigureAuth()
	}, store, cli.Args())
}
