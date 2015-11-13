package commands

import (
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
)

func cmdUpgrade(cli CommandLine, store rpcdriver.Store) error {
	return runActionOnHosts(func(h *host.Host) error {
		return h.Upgrade()
	}, store, cli.Args())
}
