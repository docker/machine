package commands

import (
	"github.com/docker/machine/libmachine/persist"
)

func cmdStop(c CommandLine, store persist.Store) error {
	return runActionWithContext("stop", c, store)
}
