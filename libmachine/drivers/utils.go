package drivers

import (
	"fmt"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
)

func GetSSHClientFromDriver(d Driver) (ssh.Client, error) {
	return GetSSHClientFromDriverWithOptions(d, &ssh.Options{})
}

func GetSSHClientFromDriverWithOptions(d Driver, options *ssh.Options) (ssh.Client, error) {
	address, err := d.GetSSHHostname()
	if err != nil {
		return nil, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return nil, err
	}

	if d.GetSSHKeyPath() != "" {
		options.Keys = []string{d.GetSSHKeyPath()}
	}

	client, err := ssh.NewClient(d.GetSSHUsername(), address, port, options)
	return client, err

}

func RunSSHCommandFromDriver(d Driver, command string) (string, error) {
	return RunSSHCommandFromDriverWithOptions(d, command, &ssh.Options{})
}

func RunSSHCommandFromDriverWithOptions(d Driver, command string, options *ssh.Options) (string, error) {
	client, err := GetSSHClientFromDriverWithOptions(d, options)
	if err != nil {
		return "", err
	}

	log.Debugf("About to run SSH command:\n%s", command)

	output, err := client.Output(command)
	log.Debugf("SSH cmd err, output: %v: %s", err, output)
	if err != nil {
		return "", fmt.Errorf(`Something went wrong running an SSH command!
command : %s
err     : %v
output  : %s
`, command, err, output)
	}

	return output, nil
}

func sshAvailableFunc(d Driver, options *ssh.Options) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForSSH function...")
		if _, err := RunSSHCommandFromDriverWithOptions(d, "exit 0", options); err != nil {
			log.Debugf("Error getting ssh command 'exit 0' : %s", err)
			return false
		}
		return true
	}
}

func WaitForSSH(d Driver) error {
	return waitForSSHWithOptions(d, &ssh.Options{})
}

func waitForSSHWithOptions(d Driver, options *ssh.Options) error {
	// Try to dial SSH for 30 seconds before timing out.
	if err := mcnutils.WaitFor(sshAvailableFunc(d, options)); err != nil {
		return fmt.Errorf("Too many retries waiting for SSH to be available.  Last error: %s", err)
	}
	return nil
}
