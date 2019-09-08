// +build !windows

package commands

import (
	"os"

	"github.com/docker/machine/libmachine"
)

func cmdForward(c CommandLine, api libmachine.API) error {
	args := c.Args()
	if len(args) < 1 || len(args) > 2 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	target, err := targetHost(c, api)
	if err != nil {
		return err
	}

	address := args[0]
	if len(args) > 1 {
		address = args[1]
	}

	hostInfoLoader := &storeHostInfoLoader{api}

	cmd, err := getForwardCmd(target, address, hostInfoLoader)
	if err != nil {
		return err
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
