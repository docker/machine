package commands

import (
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
)

func cmdIP(cli CommandLine, store rpcdriver.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.PrintIP()
	}, store, cli.Args())
}
