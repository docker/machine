package commands

import (
	"fmt"
	"os/exec"

	"github.com/docker/machine/libmachine/log"
)

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
