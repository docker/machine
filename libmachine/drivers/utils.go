package drivers

import (
	"bytes"
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/mitchellh/packer/communicator/winrm"
	"github.com/mitchellh/packer/packer"
)

func GetSSHClientFromDriver(d Driver) (ssh.Client, error) {
	address, err := d.GetSSHHostname()
	if err != nil {
		return nil, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return nil, err
	}

	var auth *ssh.Auth
	if d.GetSSHKeyPath() == "" {
		auth = &ssh.Auth{}
	} else {
		auth = &ssh.Auth{
			Keys: []string{d.GetSSHKeyPath()},
		}
	}

	client, err := ssh.NewClient(d.GetSSHUsername(), address, port, auth)
	return client, err

}

func RunSSHCommandFromDriver(d Driver, command string) (string, error) {
	client, err := GetSSHClientFromDriver(d)
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

func sshAvailableFunc(d Driver) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForSSH function...")
		if _, err := RunSSHCommandFromDriver(d, "exit 0"); err != nil {
			log.Debugf("Error getting ssh command 'exit 0' : %s", err)
			return false
		}
		return true
	}
}

func WaitForSSH(d Driver) error {
	if d.GetOS() == WINDOWS {
		return nil
	}
	// Try to dial SSH for 30 seconds before timing out.
	if err := mcnutils.WaitFor(sshAvailableFunc(d)); err != nil {
		return fmt.Errorf("Too many retries waiting for SSH to be available.  Last error: %s", err)
	}
	return nil
}

//
// WinRMUpload uploads data to a file on a Windows Server. Uses WinRM, which
// should be enabled on the target server
//
// Parameters:
//   host: target Windows server host
//   username: username for the host
//   password: password for the host
//   indata: data to be sent
//   outfile: target file to write to. the path should be specified in Windows
//      filepath format
// Returns:
//   error: errors from establishing connection and copying the file
//
func WinRMUpload(host string, username string, password string, indata string, outfile string) error {
	log.Debug("Connecting to ", username, "@", password, ":", host, " copying to ", outfile)
	c, err := winrm.New(&winrm.Config{
		Host:     host,
		Port:     5985,
		Username: username,
		Password: password,
		Timeout:  30 * time.Second,
	})

	if err != nil {
		return err
	}

	err = c.Upload(outfile, bytes.NewReader([]byte(indata)), nil)
	if err != nil {
		return err
	}
	return nil
}

//
// WinRMRunCmd runs a command on a Windows Server. Uses WinRM, which
// should be enabled on the target server
//
// Parameters:
//   host: target Windows server host
//   username: username for the host
//   password: password for the host
//   command: command to run
// Returns:
//   error: errors from establishing connection and copying the file
//
func WinRMRunCmd(host string, username string, password string, command string) (string, error) {
	log.Debug("Connecting to ", username, "@", password, ":", host, " running command ", command)
	c, err := winrm.New(&winrm.Config{
		Host:     host,
		Port:     5985,
		Username: username,
		Password: password,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		return "", err
	}

	var cmd packer.RemoteCmd
	stdout := new(bytes.Buffer)
	cmd.Command = command
	cmd.Stdout = stdout

	err = c.Start(&cmd)
	if err != nil {
		return "", err
	}
	cmd.Wait()
	return stdout.String(), nil
}
