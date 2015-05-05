package commands

import (
	"io"
	"os"
	"strings"

	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
)

func cmdSsh(c *cli.Context) {
	var (
		output ssh.Output
		err    error
	)

	if len(c.Args()) == 0 {
		log.Fatal("Error: Please specify a machine name.")
	}

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

	if len(c.Args()) == 1 {
		err = host.CreateSSHShell()
	} else {
		var (
			cmd  string
			args []string = c.Args()
		)

		for i, arg := range args {
			if arg == "--" {
				i++
				cmd = strings.Join(args[i:], " ")
				break
			}
		}
		if len(cmd) == 0 {
			cmd = strings.Join(args[1:], " ")
		}
		output, err = host.RunSSHCommand(cmd)

		io.Copy(os.Stderr, output.Stderr)
		io.Copy(os.Stdout, output.Stdout)
	}

	if err != nil {
		log.Fatal(err)
	}
}
