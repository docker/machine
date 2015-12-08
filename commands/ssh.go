package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
)

func cmdSSH(c CommandLine, api libmachine.API) error {
	// Check for help flag -- Needed due to SkipFlagParsing
	firstArg := c.Args().First()
	if firstArg == "-help" || firstArg == "--help" || firstArg == "-h" {
		c.ShowHelp()
		return nil
	}

	name := firstArg
	if name == "" {
		return ErrExpectedOneMachine
	}

	host, err := api.Load(name)
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
