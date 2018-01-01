package commands

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetForwardCmd(t *testing.T) {
	hostInfoLoader := MockHostInfoLoader{MockHostInfo{
		ip:          "12.34.56.78",
		sshPort:     234,
		sshUsername: "root",
		sshKeyPath:  "/fake/keypath/id_rsa",
	}}

	cmd, err := getForwardCmd("myfunhost", "8080::8080", &hostInfoLoader)

	expectedArgs := append(
		baseSSHArgs,
		"-o",
		"IdentitiesOnly=yes",
		"-p",
		"234",
		"-o",
		"IdentityFile=\"/fake/keypath/id_rsa\"",
		"root@12.34.56.78",
		"-f",
		"-o",
		"ExitOnForwardFailure=yes",
		"-nNTL",
		"8080::8080",
	)

	cmdPath, err := exec.LookPath("ssh")
	expectedCmd := exec.Command(cmdPath, expectedArgs...)

	assert.Equal(t, expectedCmd, cmd)
	assert.NoError(t, err)
}
