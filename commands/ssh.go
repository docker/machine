package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/state"
)

func cmdSSH(cli CommandLine, store rpcdriver.Store) error {
	// Check for help flag -- Needed due to SkipFlagParsing
	for _, arg := range cli.Args() {
		if arg == "-help" || arg == "--help" || arg == "-h" {
			cli.ShowHelp()
			return nil
		}
	}

	name := cli.Args().First()
	if name == "" {
		return ErrExpectedOneMachine
	}

	host, err := store.Load(name)
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

	return client.Shell(cli.Args().Tail()...)
}
