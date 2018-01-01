package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
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

func getForwardCmd(target, address string, hostInfoLoader HostInfoLoader) (*exec.Cmd, error) {
	var cmdPath string
	var err error

	cmdPath, err = exec.LookPath("ssh")
	if err != nil {
		return nil, err
	}

	h, err := hostInfoLoader.load(target)
	if err != nil {
		return nil, err
	}

	sshArgs := baseSSHArgs
	if h.GetSSHKeyPath() != "" {
		sshArgs = append(sshArgs, "-o", "IdentitiesOnly=yes")
	}

	port, err := h.GetSSHPort()
	if err == nil && port > 0 {
		sshArgs = append(sshArgs, "-p", fmt.Sprintf("%d", port))
	}

	if h.GetSSHKeyPath() != "" {
		sshArgs = append(sshArgs, "-o", fmt.Sprintf("IdentityFile=%q", h.GetSSHKeyPath()))
	}

	user := h.GetSSHUsername()
	hostname, err := h.GetSSHHostname()
	if err != nil {
		return nil, err
	}

	location := fmt.Sprintf("%s@%s", user, hostname)
	sshArgs = append(sshArgs, location)

	sshArgs = append(sshArgs, "-f")
	sshArgs = append(sshArgs, "-o", "ExitOnForwardFailure=yes")

	sshArgs = append(sshArgs, "-nNTL")
	sshArgs = append(sshArgs, address)

	cmd := exec.Command(cmdPath, sshArgs...)
	log.Debug(*cmd)
	return cmd, nil
}
