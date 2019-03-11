package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/machine/commands/commandstest"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

const (
	trustedCertsFolder = "/var/lib/boot2docker/certs"
)

var (
	errFirstArgIsNotPEMFile = errors.New("The first argument should the path of a file with extention '.pem'")
)

func cmdAddCACertificate(c CommandLine, api libmachine.API) error {
	args := c.Args()
	if len(args) == 0 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	if len(args) <= 2 && !strings.Contains(args[0], ".pem") {
		return errFirstArgIsNotPEMFile
	}

	target, err := targetHostWithOffset(c, api, 1)
	if err != nil {
		return err
	}

	err = createFolder(target, api)
	if err != nil {
		return err
	}

	src := args[0]
	targetPath := fmt.Sprintf("%s:%s", target, trustedCertsFolder)

	err = uploadCertificate(src, targetPath, api)
	if err != nil {
		return err
	}

	if c.Bool("restart") {
		if err := runAction("restart", c, api); err != nil {
			return err
		}
	} else {
		log.Info("In order for the change to be effective, you need to restart the docker-machine with:\n%s", "docker-machine restart")
	}

	return nil
}

func createFolder(targetHost string, api libmachine.API) error {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: strings.Split(fmt.Sprintf("sudo mkdir -p %s", trustedCertsFolder), " "),
	}

	log.Debug(commandLine)
	err := cmdSSH(commandLine, api)
	if err != nil {
		return err
	}
	return nil
}

func uploadCertificate(src, targetPath string, api libmachine.API) error {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{
			src,
			targetPath,
		},
	}
	log.Debug(commandLine)
	err := cmdScp(commandLine, api)
	if err != nil {
		return err
	}
	log.Info("Certificate uploaded")
	return nil
}
