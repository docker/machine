package ssh

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/pkg/term"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type Client interface {
	Output(command string) (string, error)
	Shell(args ...string) error
}

type ExternalClient struct {
	BaseArgs   []string
	BinaryPath string
}

type NativeClient struct {
	Config   ssh.ClientConfig
	Hostname string
	Port     int
}

type Auth struct {
	Passwords []string
	Keys      []string
}

type ClientType string

const (
	maxDialAttempts = 10
)

const (
	External ClientType = "external"
	Native   ClientType = "native"
)

var (
	baseSSHArgs = []string{
		"-o", "BatchMode=yes",
		"-o", "PasswordAuthentication=no",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=quiet", // suppress "Warning: Permanently added '[localhost]:2022' (ECDSA) to the list of known hosts."
		"-o", "ConnectionAttempts=3", // retry 3 times if SSH connection fails
		"-o", "ConnectTimeout=10", // timeout after 10 seconds
		"-o", "ControlMaster=no", // disable ssh multiplexing
		"-o", "ControlPath=none",
	}
	defaultClientType = External
)

func SetDefaultClient(clientType ClientType) {
	// Allow over-riding of default client type, so that even if ssh binary
	// is found in PATH we can still use the Go native implementation if
	// desired.
	switch clientType {
	case External:
		defaultClientType = External
	case Native:
		defaultClientType = Native
	}
}

func NewClient(user string, host string, port int, auth *Auth) (Client, error) {
	sshBinaryPath, err := exec.LookPath("ssh")
	if err != nil {
		log.Debug("SSH binary not found, using native Go implementation")
		client, err := NewNativeClient(user, host, port, auth)
		log.Debug(client)
		return client, err
	}

	if defaultClientType == Native {
		log.Debug("Using SSH client type: native")
		client, err := NewNativeClient(user, host, port, auth)
		log.Debug(client)
		return client, err
	}

	log.Debug("Using SSH client type: external")
	client, err := NewExternalClient(sshBinaryPath, user, host, port, auth)
	log.Debug(client)
	return client, err
}

func NewNativeClient(user, host string, port int, auth *Auth) (Client, error) {
	config, err := NewNativeConfig(user, auth)
	if err != nil {
		return nil, fmt.Errorf("Error getting config for native Go SSH: %s", err)
	}

	return NativeClient{
		Config:   config,
		Hostname: host,
		Port:     port,
	}, nil
}

func NewNativeConfig(user string, auth *Auth) (ssh.ClientConfig, error) {
	var (
		authMethods []ssh.AuthMethod
	)

	for _, k := range auth.Keys {
		key, err := ioutil.ReadFile(k)
		if err != nil {
			return ssh.ClientConfig{}, err
		}

		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return ssh.ClientConfig{}, err
		}

		authMethods = append(authMethods, ssh.PublicKeys(privateKey))
	}

	for _, p := range auth.Passwords {
		authMethods = append(authMethods, ssh.Password(p))
	}

	return ssh.ClientConfig{
		User: user,
		Auth: authMethods,
	}, nil
}

func (client NativeClient) dialSuccess() bool {
	if _, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), &client.Config); err != nil {
		log.Debugf("Error dialing TCP: %s", err)
		return false
	}
	return true
}

func (client NativeClient) session(command string) (*ssh.Session, error) {
	if err := mcnutils.WaitFor(client.dialSuccess); err != nil {
		return nil, fmt.Errorf("Error attempting SSH client dial: %s", err)
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), &client.Config)
	if err != nil {
		return nil, fmt.Errorf("Mysterious error dialing TCP for SSH (we already succeeded at least once) : %s", err)
	}

	return conn.NewSession()
}

func (client NativeClient) Output(command string) (string, error) {
	session, err := client.session(command)
	if err != nil {
		return "", nil
	}

	output, err := session.CombinedOutput(command)
	defer session.Close()

	return string(output), err
}

func (client NativeClient) OutputWithPty(command string) (string, error) {
	session, err := client.session(command)
	if err != nil {
		return "", nil
	}

	fd := int(os.Stdin.Fd())

	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return "", err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// request tty -- fixes error with hosts that use
	// "Defaults requiretty" in /etc/sudoers - I'm looking at you RedHat
	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return "", err
	}

	output, err := session.CombinedOutput(command)
	defer session.Close()

	return string(output), err
}

func (client NativeClient) Shell(args ...string) error {
	var (
		termWidth, termHeight int
	)
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), &client.Config)
	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	fd := os.Stdin.Fd()

	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}

		defer term.RestoreTerminal(fd, oldState)

		winsize, err := term.GetWinsize(fd)
		if err != nil {
			termWidth = 80
			termHeight = 24
		} else {
			termWidth = int(winsize.Width)
			termHeight = int(winsize.Height)
		}
	}

	if err := session.RequestPty("xterm", termHeight, termWidth, modes); err != nil {
		return err
	}

	if len(args) == 0 {
		if err := session.Shell(); err != nil {
			return err
		}
		session.Wait()
	} else {
		session.Run(strings.Join(args, " "))
	}

	return nil
}

func NewExternalClient(sshBinaryPath, user, host string, port int, auth *Auth) (ExternalClient, error) {
	client := ExternalClient{
		BinaryPath: sshBinaryPath,
	}

	args := append(baseSSHArgs, fmt.Sprintf("%s@%s", user, host))

	// If no identities are explicitly provided, also look at the identities
	// offered by ssh-agent
	if len(auth.Keys) > 0 {
		args = append(args, "-o", "IdentitiesOnly=yes")
	}

	// Specify which private keys to use to authorize the SSH request.
	for _, privateKeyPath := range auth.Keys {
		if privateKeyPath != "" {
			args = append(args, "-i", privateKeyPath)
		}
	}

	// Set which port to use for SSH.
	args = append(args, "-p", fmt.Sprintf("%d", port))

	client.BaseArgs = args

	return client, nil
}

func getSSHCmd(binaryPath string, args ...string) *exec.Cmd {
	return exec.Command(binaryPath, args...)
}

func (client ExternalClient) Output(command string) (string, error) {
	args := append(client.BaseArgs, command)
	cmd := getSSHCmd(client.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (client ExternalClient) Shell(args ...string) error {
	args = append(client.BaseArgs, args...)
	cmd := getSSHCmd(client.BinaryPath, args...)

	log.Debug(cmd)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
