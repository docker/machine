package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

func cmdRestart(c CommandLine, store persist.Store) error {
	if err := runActionOnHosts(func(h *host.Host) error {
		return h.Restart()
	}, store, c.Args()); err != nil {
		return err
	}

	log.Info("Restarted machines may have new IP addresses. You may need to re-run the `docker-machine env` command.")

	return nil
}
