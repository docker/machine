package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
)

func cmdUpgrade(cli CommandLine, store persist.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.Upgrade()
	}, store, cli.Args())
}
