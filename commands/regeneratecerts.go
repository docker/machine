package commands

import (
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
)

func cmdRegenerateCerts(cli CommandLine, store rpcdriver.Store) error {
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
