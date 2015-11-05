package commands

import (
	"github.com/docker/machine/libmachine/persist"
)

func cmdUpgrade(c CommandLine, store persist.Store) error {
	return runActionWithContext("upgrade", c, store)
}
