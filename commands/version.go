package commands

import "github.com/docker/machine/libmachine"

func cmdVersion(c CommandLine, api libmachine.API) error {
	c.ShowVersion()
	return nil
}
