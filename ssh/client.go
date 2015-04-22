package ssh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/pkg/term"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Config   *ssh.ClientConfig
	Hostname string
	Port     int
}

func NewClient(user string, host string, port int, auth *Auth) (*Client, error) {
	config, err := NewConfig(user, auth)
	if err != nil {
		return nil, err
	}

	return &Client{
		Config:   config,
		Hostname: host,
		Port:     port,
	}, nil
}

func NewConfig(user string, auth *Auth) (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	for _, k := range auth.Keys {
		key, err := ioutil.ReadFile(k)
		if err != nil {
			return nil, err
		}

		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		authMethods = append(authMethods, ssh.PublicKeys(privateKey))
	}

	for _, p := range auth.Passwords {
		authMethods = append(authMethods, ssh.Password(p))
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: authMethods,
	}, nil
}

func (client *Client) Run(command string) (Output, error) {
	var output Output

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), client.Config)
	if err != nil {
		return output, err
	}

	session, err := conn.NewSession()
	if err != nil {
		return output, err
	}

	defer session.Close()

	var stdout, stderr bytes.Buffer

	session.Stdout = &stdout
	session.Stderr = &stderr

	output = Output{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	return output, session.Run(command)
}

func (client *Client) Shell() error {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), client.Config)
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

	var termWidth, termHeight int

	fd := os.Stdin.Fd()

	if term.IsTerminal(fd) {
		var oldState *term.State

		oldState, err = term.MakeRaw(fd)
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

	if err := session.Shell(); err != nil {
		return err
	}

	session.Wait()

	return nil
}

type Auth struct {
	Passwords []string
	Keys      []string
}

type Output struct {
	Stdout io.Reader
	Stderr io.Reader
}
