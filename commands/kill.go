package commands

import (
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/host"
)

func cmdKill(cli CommandLine, store persist.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.Kill()
	}, store, cli.Args())
}
