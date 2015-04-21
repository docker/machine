package commands

import (
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
)

func cmdSsh(c *cli.Context) {
	var (
		err    error
		sshCmd *exec.Cmd
	)
	name := c.Args().First()

	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	mcn, err := newMcn(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	if name == "" {
		host, err := mcn.GetActive()
		if err != nil {
			log.Fatalf("unable to get active host: %v", err)
		}

		if host == nil {
			log.Fatalf("There is no active host. Please set it with %s active <machine name>.", c.App.Name)
		}

		name = host.Name
	}

	host, err := mcn.Get(name)
	if err != nil {
		log.Fatal(err)
	}

	_, err = host.GetURL()
	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			log.Fatalf("%s is not running. Please start this with docker-machine start %s", host.Name, host.Name)
		} else {
			log.Fatalf("Unexpected error getting machine url: %s", err)
		}
	}

	if len(c.Args()) <= 1 {
		sshCmd, err = host.GetSSHCommand()
	} else {
		sshCmd, err = host.GetSSHCommand(c.Args()[1:]...)
	}
	if err != nil {
		log.Fatal(err)
	}

	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		log.Fatal(err)
	}
}
