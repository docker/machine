package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/docker/machine/commands/commandstest"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

const (
	trustedCertsFolder = "/var/lib/boot2docker/certs/"
)

var (
	errFirstArgIsNotPEMFile = errors.New("The first argument should the path of a file with extention '.pem'")
)

func cmdAddCACertificate(c CommandLine, api libmachine.API) error {
	stdout := os.Stdout
	args := c.Args()
	if len(args) == 0 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	if len(args) <= 2 && !strings.Contains(args[0], ".pem") {
		return errFirstArgIsNotPEMFile
	}

	target, err := targetHost(c, api, 1)
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
	//restoring the stdout captured by ssh/scp
	os.Stdout = stdout

	if c.Bool("restart") {
		if err := restart(target, api); err != nil {
			return err
		}
	} else {
		fmt.Printf("In order for the change to be effective, you need to restart the docker-machine with:\n%s", "docker-machine restart")
	}

	return nil
}

func createFolder(targetHost string, api libmachine.API) error {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{
			targetHost,
			fmt.Sprintf("sudo mkdir -p %s", trustedCertsFolder),
			" && ",
			fmt.Sprintf("sudo chown docker:docker %s", trustedCertsFolder),
		},
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
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"quiet": true,
			},
		},
	}

	log.Debug(commandLine)
	err := cmdScp(commandLine, api)
	if err != nil {
		return err
	}
	fmt.Println("Certificate uploaded successfully")
	return nil
}

func restart(targetHost string, api libmachine.API) error {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{
			targetHost,
		},
	}

	log.Debug(commandLine)
	if err := runAction("restart", commandLine, api); err != nil {
		return err
	}

	return nil
}
