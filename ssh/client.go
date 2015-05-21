package ssh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/docker/docker/pkg/term"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Config   *ssh.ClientConfig
	Hostname string
	Port     int
}

const (
	maxDialAttempts = 10
)

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

func dialSuccess(client *Client) func() bool {
	return func() bool {
		if _, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), client.Config); err != nil {
			log.Debugf("Error dialing TCP: %s", err)
			return false
		}
		return true
	}
}

func (client *Client) Run(command string) (Output, error) {
	var (
		output         Output
		stdout, stderr bytes.Buffer
	)

	if err := utils.WaitFor(dialSuccess(client)); err != nil {
		return output, fmt.Errorf("Error attempting SSH client dial: %s", err)
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Hostname, client.Port), client.Config)
	if err != nil {
		return output, fmt.Errorf("Mysterious error dialing TCP for SSH (we already succeeded at least once) : %s", err)
	}

	session, err := conn.NewSession()
	if err != nil {
		return output, fmt.Errorf("Error getting new session: %s", err)
	}

	defer session.Close()

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

	go func() {
		for {
			time.Sleep(time.Millisecond * 250)
			h, w := client.getWinsize(fd)

			payload := ssh.Marshal(&winsizeReq{
				WidthCol:  uint32(w),
				HeightCol: uint32(h),
				WidthPx:   uint32(w * 8),
				HeightPx:  uint32(h * 8),
			})

			session.SendRequest("window-change", false, payload)
		}
	}()

	session.Wait()

	return nil
}

func (client *Client) getWinsize(fd uintptr) (int, int) {
	ws, err := term.GetWinsize(fd)
	if err != nil {
		if ws == nil {
			return 24, 80
		}
	}
	return int(ws.Height), int(ws.Width)
}

type Auth struct {
	Passwords []string
	Keys      []string
}

type Output struct {
	Stdout io.Reader
	Stderr io.Reader
}

type winsizeReq struct {
	WidthCol  uint32
	HeightCol uint32
	WidthPx   uint32
	HeightPx  uint32
}
