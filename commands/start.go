package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

func cmdStart(cli CommandLine, store persist.Store) error {
	if err := runActionOnHosts(func(h *host.Host) error {
		return h.Start()
	}, store, cli.Args()); err != nil {
		return err
	}

	log.Info("Started machines may have new IP addresses. You may need to re-run the `docker-machine env` command.")

	return nil
}
