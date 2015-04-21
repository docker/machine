package ssh

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func GetSSHCommand(host string, port int, user string, sshKey string, args ...string) *exec.Cmd {
	defaultSSHArgs := []string{
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=no", // don't bother checking in ~/.ssh/known_hosts
		"-o", "UserKnownHostsFile=/dev/null", // don't write anything to ~/.ssh/known_hosts
		"-o", "ConnectionAttempts=3", // retry 3 times if SSH connection fails
		"-o", "ConnectTimeout=10", // timeout after 10 seconds
		"-o", "LogLevel=quiet", // suppress "Warning: Permanently added '[localhost]:2022' (ECDSA) to the list of known hosts."
		"-p", fmt.Sprintf("%d", port),
		"-i", sshKey,
		fmt.Sprintf("%s@%s", user, host),
	}

	sshArgs := append(defaultSSHArgs, args...)
	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stderr = os.Stderr

	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
	}

	log.Debugf("executing: %v", strings.Join(cmd.Args, " "))

	return cmd
}

func WaitForTCP(addr string) error {
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			continue
		}
		defer conn.Close()
		if _, err = conn.Read(make([]byte, 1)); err != nil {
			continue
		}
		break
	}
	return nil
}
