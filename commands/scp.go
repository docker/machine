package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/log"
)

var (
	ErrMalformedInput = fmt.Errorf("The input was malformed")
)

var (
	// TODO: possibly move this to ssh package
	baseSSHArgs = []string{
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=quiet", // suppress "Warning: Permanently added '[localhost]:2022' (ECDSA) to the list of known hosts."
	}
)

func getInfoForScpArg(hostAndPath string, provider libmachine.Provider) (*libmachine.Host, string, []string, error) {
	// TODO: What to do about colon in filepath?
	splitInfo := strings.Split(hostAndPath, ":")

	// Host path.  e.g. "/tmp/foo"
	if len(splitInfo) == 1 {
		return nil, splitInfo[0], nil, nil
	}

	// Remote path.  e.g. "machinename:/usr/bin/cmatrix"
	if len(splitInfo) == 2 {
		path := splitInfo[1]
		host, err := provider.Get(splitInfo[0])
		if err != nil {
			return nil, "", nil, fmt.Errorf("Error loading host: %s", err)
		}
		args := []string{
			"-i",
			host.Driver.GetSSHKeyPath(),
		}
		return host, path, args, nil
	}

	return nil, "", nil, ErrMalformedInput
}

func generateLocationArg(host *libmachine.Host, path string) (string, error) {
	locationPrefix := ""
	if host != nil {
		ip, err := host.Driver.GetIP()
		if err != nil {
			return "", err
		}
		locationPrefix = fmt.Sprintf("%s@%s:", host.Driver.GetSSHUsername(), ip)
	}
	return locationPrefix + path, nil
}

func getScpCmd(src, dest string, sshArgs []string, provider libmachine.Provider) (*exec.Cmd, error) {
	cmdPath, err := exec.LookPath("scp")
	if err != nil {
		return nil, errors.New("Error: You must have a copy of the scp binary locally to use the scp feature.")
	}

	srcHost, srcPath, srcOpts, err := getInfoForScpArg(src, provider)
	if err != nil {
		return nil, err
	}

	destHost, destPath, destOpts, err := getInfoForScpArg(dest, provider)
	if err != nil {
		return nil, err
	}

	// Append needed -i / private key flags to command.
	sshArgs = append(sshArgs, srcOpts...)
	sshArgs = append(sshArgs, destOpts...)

	// Append actual arguments for the scp command (i.e. docker@<ip>:/path)
	locationArg, err := generateLocationArg(srcHost, srcPath)
	if err != nil {
		return nil, err
	}
	sshArgs = append(sshArgs, locationArg)
	locationArg, err = generateLocationArg(destHost, destPath)
	if err != nil {
		return nil, err
	}
	sshArgs = append(sshArgs, locationArg)

	cmd := exec.Command(cmdPath, sshArgs...)
	log.Debug(*cmd)
	return cmd, nil
}

func runCmdWithStdIo(cmd exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func cmdScp(c *cli.Context) {
	args := c.Args()
	if len(args) != 2 {
		cli.ShowCommandHelp(c, "scp")
		log.Fatal("Improper number of arguments.")
	}

	// TODO: Check that "-3" flag is available in user's version of scp.
	// It is on every system I've checked, but the manual mentioned it's "newer"
	sshArgs := append(baseSSHArgs, "-3")

	if c.Bool("recursive") {
		sshArgs = append(sshArgs, "-r")
	}

	src := args[0]
	dest := args[1]

	provider := getDefaultProvider(c)
	cmd, err := getScpCmd(src, dest, sshArgs, *provider)

	if err != nil {
		log.Fatal(err)
	}
	if err := runCmdWithStdIo(*cmd); err != nil {
		log.Fatal(err)
	}
}
