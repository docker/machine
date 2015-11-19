package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
)

func cmdIP(c CommandLine, store persist.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.PrintIP()
	}, store, c.Args())
}
