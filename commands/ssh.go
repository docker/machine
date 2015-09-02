package commands

import (
	"github.com/docker/machine/cli"
	"github.com/docker/machine/libmachine/state"
)

func cmdSsh(c *cli.Context) {
	args := c.Args()
	name := args.First()

	if name == "" {
		fatal("Error: Please specify a machine name.")
	}

	store := getStore(c)
	host, err := loadHost(store, name)
	if err != nil {
		fatal(err)
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		fatal(err)
	}

	if currentState != state.Running {
		fatalf("Error: Cannot run SSH command: Host %q is not running", host.Name)
	}

	client, err := host.CreateSSHClient()
	if err != nil {
		fatal(err)
	}

	if err := client.Shell(c.Args().Tail()...); err != nil {
		fatal(err)
	}
}
