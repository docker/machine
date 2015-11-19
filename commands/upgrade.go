package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
)

func cmdUpgrade(c CommandLine, store persist.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.Upgrade()
	}, store, c.Args())
}
