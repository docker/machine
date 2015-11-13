package commands

import (
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
)

func cmdStop(cli CommandLine, store rpcdriver.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.Stop()
	}, store, cli.Args())
}
