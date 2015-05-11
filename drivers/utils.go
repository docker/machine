package drivers

import (
	"fmt"

	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/utils"
)

func RunSSHCommandFromDriver(d Driver, command string) (ssh.Output, error) {
	var output ssh.Output

	addr, err := d.GetSSHHostname()
	if err != nil {
		return output, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return output, err
	}

	auth := &ssh.Auth{
		Keys: []string{d.GetSSHKeyPath()},
	}

	client, err := ssh.NewClient(d.GetSSHUsername(), addr, port, auth)
	if err != nil {
		return output, err
	}

	output, err = client.Run(command)
	log.Debug(fmt.Sprintf("SSH cmd err, output: %v: %s", err, output))
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
	if err := utils.WaitFor(sshAvailableFunc(d)); err != nil {
		return fmt.Errorf("Too many retries.  Last error: %s", err)
	}
	return nil
}
