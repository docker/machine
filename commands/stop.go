package commands

import (
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
)

func cmdStop(cli CommandLine, store persist.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.Stop()
	}, store, cli.Args())
}
