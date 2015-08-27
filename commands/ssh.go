package commands

import (
	"fmt"
	"strings"

	"github.com/docker/machine/log"
	"github.com/docker/machine/state"

	"github.com/codegangsta/cli"
)

func cmdSsh(c *cli.Context) {
	args := c.Args()
	name := args.First()

	if name == "" {
		log.Fatal("Error: Please specify a machine name.")
	}

	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := newProvider(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	host, err := provider.Get(name)
	if err != nil {
		log.Fatal(err)
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Fatal(err)
	}

	if currentState != state.Running {
		log.Fatalf("Error: Cannot run SSH command: Host %q is not running", host.Name)
	}

	if len(c.Args()) == 1 {
		err := host.CreateSSHShell()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		output, err := host.RunSSHCommand(strings.Join(c.Args().Tail(), " "))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print(output)
	}

}
