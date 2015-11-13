package commands

import "github.com/docker/machine/libmachine/persist"

func cmdUpgrade(c CommandLine, store persist.Store) error {
	return runActionOnHosts("upgrade", store, c.Args())
}
