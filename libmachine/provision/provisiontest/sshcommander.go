//Package provisiontest provides utilities for testing provisioners
package provisiontest

import (
	"errors"
	"strings"
)

//FakeSSHCommander is an implementation of provision.SSHCommander to provide predictable responses set by testing code
//Extend it when needed
type FakeSSHCommander struct {
	//Result of the ssh command to look up the FilesystemType
	FilesystemType string
}

//SSHCommand is an implementation of provision.SSHCommander.SSHCommand to provide predictable responses set by testing code
func (sshCmder FakeSSHCommander) SSHCommand(args string) (string, error) {
	if !strings.HasPrefix(args, "stat -f") {
		return "", errors.New("Not implemented by FakeSSHCommander")
	}
	return sshCmder.FilesystemType + "\n", nil
}
