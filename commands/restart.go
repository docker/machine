package commands

import (
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
)

func cmdRestart(cli CommandLine, store rpcdriver.Store) error {
	if err := runActionOnHosts(func(h *host.Host) error {
		return h.Restart()
	}, store, cli.Args()); err != nil {
		return err
	}

	log.Info("Restarted machines may have new IP addresses. You may need to re-run the `docker-machine env` command.")

	return nil
}
