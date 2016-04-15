package ssh

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSSHCmdArgs(t *testing.T) {
	cases := []struct {
		binaryPath   string
		args         []string
		expectedArgs []string
	}{
		{
			binaryPath: "/usr/local/bin/ssh",
			args: []string{
				"docker@localhost",
				"apt-get install -y htop",
			},
			expectedArgs: []string{
				"/usr/local/bin/ssh",
				"docker@localhost",
				"apt-get install -y htop",
			},
		},
		{
			binaryPath: "C:\\Program Files\\Git\\bin\\ssh.exe",
			args: []string{
				"docker@localhost",
				"sudo /usr/bin/sethostname foobar && echo 'foobar' | sudo tee /var/lib/boot2docker/etc/hostname",
			},
			expectedArgs: []string{
				"C:\\Program Files\\Git\\bin\\ssh.exe",
				"docker@localhost",
				"sudo /usr/bin/sethostname foobar && echo 'foobar' | sudo tee /var/lib/boot2docker/etc/hostname",
			},
		},
	}

	for _, c := range cases {
		cmd := getSSHCmd(c.binaryPath, c.args...)
		assert.Equal(t, cmd.Args, c.expectedArgs)
	}
}

func TestNewExternalClient(t *testing.T) {
	keyFile, err := ioutil.TempFile("", "docker-machine-tests-dummy-private-key")
	if err != nil {
		t.Fatal(err)
	}
	defer keyFile.Close()

	keyFilename := keyFile.Name()
	defer os.Remove(keyFilename)

	cases := []struct {
		sshBinaryPath string
		user          string
		host          string
		port          int
		perm          os.FileMode
		keys          []string
		configFile    string
		expectedError string
		skipOS        string
	}{
		{
			sshBinaryPath: "/usr/local/bin/ssh",
			user:          "docker",
			host:          "localhost",
			port:          22,
			keys:          []string{"/tmp/private-key-not-exist"},
			configFile:    "/dev/null",
			expectedError: "stat /tmp/private-key-not-exist: no such file or directory",
			skipOS:        "none",
		},
		{
			keys:       []string{keyFilename},
			configFile: "/dev/null",
			perm:       0400,
			skipOS:     "windows",
		},
		{
			keys:          []string{keyFilename},
			configFile:    "/dev/null",
			perm:          0100,
			expectedError: fmt.Sprintf("'%s' is not readable", keyFilename),
			skipOS:        "windows",
		},
		{
			keys:          []string{keyFilename},
			configFile:    "/dev/null",
			perm:          0644,
			expectedError: fmt.Sprintf("permissions 0644 for '%s' are too open", keyFilename),
			skipOS:        "windows",
		},
		{
			sshBinaryPath: "/usr/local/bin/ssh",
			user:          "docker",
			host:          "localhost",
			port:          22,
			keys:          []string{},
			configFile:    "/dev/zero",
			expectedError: "",
			skipOS:        "none",
		},
		{
			sshBinaryPath: "/usr/local/bin/ssh",
			user:          "docker",
			host:          "localhost",
			port:          22,
			keys:          []string{},
			configFile:    "/tmp/does/not/exist",
			expectedError: "stat /tmp/does/not/exist: no such file or directory",
			skipOS:        "none",
		},
	}

	for _, c := range cases {
		options := &Options{
			Keys:       c.keys,
			ConfigFile: c.configFile,
		}
		if runtime.GOOS != c.skipOS {
			keyFile.Chmod(c.perm)
			cli, err := NewExternalClient(c.sshBinaryPath, c.user, c.host, c.port, options)
			if c.expectedError == "" {
				assert.Nil(t, err)
				argsStr := strings.Join(cli.BaseArgs, " ")
				assert.Contains(t, argsStr, fmt.Sprintf("-F %s", c.configFile))
			} else {
				assert.EqualError(t, err, c.expectedError)
			}
		}
	}
}
