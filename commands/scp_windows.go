package commands

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/docker/machine/libmachine"
)

func cmdScp(c CommandLine, api libmachine.API) error {
	args := c.Args()
	if len(args) != 2 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	src := args[0]
	dest := args[1]

	hostInfoLoader := &storeHostInfoLoader{api}

	cmd, err := getScpCmd(src, dest, c.Bool("recursive"), c.Bool("delta"), c.Bool("quiet"), hostInfoLoader)
	if err != nil {
		return err
	}

	// Default argument escaping is not valid for scp.exe with quoted arguments, so we do it ourselves
	// see golang/go#15566
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.CmdLine = fmt.Sprintf("\"%s\" %s", cmd.Path, strings.Join(tailArgs(cmd.Args), " "))
	return stdCommandRunner.run(*cmd)
}

func tailArgs(a []string) []string {
	if len(a) >= 2 {
		return []string(a)[1:]
	}
	return []string{}
}
