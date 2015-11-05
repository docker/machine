package commands

import (
	"github.com/docker/machine/libmachine/persist"
)

func cmdKill(c CommandLine, store persist.Store) error {
	return runActionWithContext("kill", c, store)
}
