package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
)

func cmdStop(c CommandLine, store persist.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.Stop()
	}, store, c.Args())
}
