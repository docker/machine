package commands

import (
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"

	"github.com/codegangsta/cli"
)

func cmdSsh(c *cli.Context) {
	args := c.Args()
	name := args.First()

	if name == "" {
		log.Fatal("Error: Please specify a machine name.")
	}

	store := getStore(c)
	host, err := store.Load(name)
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

	client, err := host.CreateSSHClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Shell(c.Args().Tail()...); err != nil {
		log.Fatal(err)
	}
}
