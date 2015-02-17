package main

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/ssh"
)

var (
	ErrMalformedInput = fmt.Errorf("The input was malformed")
)

func getInfoForScpArg(hostAndPath string, store Store) (*Host, string, []string, error) {
	// TODO: What to do about colon in filepath?
	splitInfo := strings.Split(hostAndPath, ":")

	// Host path.  e.g. "/tmp/foo"
	if len(splitInfo) == 1 {
		return nil, splitInfo[0], nil, nil
	}

	// Remote path.  e.g. "machinename:/usr/bin/cmatrix"
	if len(splitInfo) == 2 {
		path := splitInfo[1]
		host, err := store.Load(splitInfo[0])
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

func generateLocationArg(host *Host, path string) (string, error) {
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

func GetScpCmd(src, dest string, store Store) (*exec.Cmd, error) {
	// TODO: Check that "-3" flag is available in user's version of scp.
	// It is on every system I've checked, but the manual mentioned it's "newer"
	scpArgs := append(ssh.BaseSSHOptions, "-r", "-3")
	srcHost, srcPath, srcOpts, err := getInfoForScpArg(src, store)
	if err != nil {
		return nil, err
	}
	destHost, destPath, destOpts, err := getInfoForScpArg(dest, store)
	if err != nil {
		return nil, err
	}

	// Append needed -i / private key flags to command.
	scpArgs = append(scpArgs, srcOpts...)
	scpArgs = append(scpArgs, destOpts...)

	// Append actual arguments for the scp command (i.e. docker@<ip>:/path)
	locationArg, err := generateLocationArg(srcHost, srcPath)
	if err != nil {
		return nil, err
	}
	scpArgs = append(scpArgs, locationArg)
	locationArg, err = generateLocationArg(destHost, destPath)
	if err != nil {
		return nil, err
	}
	scpArgs = append(scpArgs, locationArg)

	cmd := exec.Command("scp", scpArgs...)
	log.Debug(*cmd)
	return cmd, nil
}
