package drivers

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
)

const (
	ErrExitCode255 = "255"
)

func GetSSHClientFromDriver(d Driver) (ssh.Client, error) {
	addr, err := d.GetSSHHostname()
	if err != nil {
		return nil, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return nil, err
	}

	auth := &ssh.Auth{
		Keys: []string{d.GetSSHKeyPath()},
	}

	client, err := ssh.NewClient(d.GetSSHUsername(), addr, port, auth)
	return client, err

}

func isErr255Exit(err error) bool {
	return strings.Contains(err.Error(), ErrExitCode255)
}

func RunSSHCommandFromDriver(d Driver, command string) (string, error) {
	client, err := GetSSHClientFromDriver(d)
	if err != nil {
		return "", err
	}

	log.Debugf("About to run SSH command:\n%s", command)

	output, err := client.Output(command)
	log.Debugf("SSH cmd err, output: %v: %s", err, output)
	if err != nil && !isErr255Exit(err) {
		log.Error("SSH cmd error!")
		log.Errorf("command: %s", command)
		log.Errorf("err    : %v", err)
		log.Fatalf("output : %s", output)
	}

	return output, err
}

func sshAvailableFunc(d Driver) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForSSH function...")
		hostname, err := d.GetSSHHostname()
		if err != nil {
			log.Debugf("Error getting IP address waiting for SSH: %s", err)
			return false
		}
		port, err := d.GetSSHPort()
		if err != nil {
			log.Debugf("Error getting SSH port: %s", err)
			return false
		}
		if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", hostname, port)); err != nil {
			log.Debugf("Error waiting for TCP waiting for SSH: %s", err)
			return false
		}

		if _, err := RunSSHCommandFromDriver(d, "exit 0"); err != nil {
			log.Debugf("Error getting ssh command 'exit 0' : %s", err)
			return false
		}
		return true
	}
}

func WaitForSSH(d Driver) error {
	if err := mcnutils.WaitFor(sshAvailableFunc(d)); err != nil {
		return fmt.Errorf("Too many retries.  Last error: %s", err)
	}
	return nil
}
