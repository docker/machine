package commands

import (
	"fmt"
	"os"
	"strings"
	"syscall"

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

	// Default argument escaping is not valid for ssh.exe with quoted arguments, so we do it ourselves
	// see golang/go#15566
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.CmdLine = fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
