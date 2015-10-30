package commands

import (
	"fmt"

	"github.com/docker/machine/cli"
	"github.com/docker/machine/libmachine/state"
)

func cmdSsh(c *cli.Context) error {
	// Check for help flag -- Needed due to SkipFlagParsing
	for _, arg := range c.Args() {
		if arg == "-help" || arg == "--help" || arg == "-h" {
			cli.ShowCommandHelp(c, "ssh")
			return nil
		}
	}

	name := c.Args().First()
	if name == "" {
		return ErrExpectedOneMachine
	}

	store := getStore(c)
	host, err := loadHost(store, name)
	if err != nil {
		return err
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return err
	}

	if currentState != state.Running {
		return fmt.Errorf("Error: Cannot run SSH command: Host %q is not running", host.Name)
	}

	client, err := host.CreateSSHClient()
	if err != nil {
		return err
	}

	return client.Shell(c.Args().Tail()...)
}
