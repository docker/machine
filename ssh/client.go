package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/docker/machine/utils"
)

var (
	ErrKnownHostsFailure = errors.New("Known host failure")
)

var (
	KnownHostsFilePath = filepath.Join(utils.GetMachineDir(), "known_hosts")
)

type Client struct {
	Host string
	Port uint16

	User       string
	PrivateKey []byte

	ClientConfig *ssh.ClientConfig

	readHandler func(filename string) ([]byte, error)
}

func NewClient(host string, port uint16, username string, privateKey []byte) *Client {
	client := &Client{
		Host: host,
		Port: port,

		User:       username,
		PrivateKey: privateKey,

		readHandler: ioutil.ReadFile,
	}

	return client
}

func (c *Client) Config() (*ssh.ClientConfig, error) {
	privateKey, err := ssh.ParsePrivateKey(c.PrivateKey)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},

		// HostKeyCallback: c.KnownHostsLookup,
	}

	return config, nil
}

func (c *Client) KnownHostsLookup(hostname string, remote net.Addr, key ssh.PublicKey) error {
	retErr := ErrKnownHostsFailure

	data, err := c.readHandler(KnownHostsFilePath)
	if err != nil {
		return err
	}

	match := strings.Trim(fmt.Sprintf(
		"%s %s",
		remote,
		string(ssh.MarshalAuthorizedKey(key)),
	), "\n")

	for _, s := range strings.Split(string(data), "\n") {
		if match == s {
			retErr = nil
		}
	}

	return retErr
}

func (c *Client) PerformCommand(args string) (io.Reader, io.Reader, error) {
	config, err := c.Config()
	if err != nil {
		return nil, nil, err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), config)
	if err != nil {
		return nil, nil, err
	}

	session, err := conn.NewSession()
	if err != nil {
		return nil, nil, err
	}

	outReader, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	errReader, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := session.Run(args); err != nil {
		return nil, nil, err
	}

	session.Close()

	return outReader, errReader, nil
}

func (c *Client) Terminal() error {
	config, err := c.Config()
	if err != nil {
		return err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), config)
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
	in, _ := session.StdinPipe()

	modes := ssh.TerminalModes{
		ssh.ECHO: 0,
	}

	if err := session.RequestPty("vt100", 256, 40, modes); err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		str, _ := reader.ReadString('\n')
		if _, err := fmt.Fprint(in, str); err != nil {
			break
		}
	}

	return nil
}
