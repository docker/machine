package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers/rpc"
)

func cmdURL(cli CommandLine, store rpcdriver.Store) error {
	if len(cli.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := store.Load(cli.Args().First())
	if err != nil {
		return err
	}

	url, err := host.GetURL()
	if err != nil {
		return err
	}

	fmt.Println(url)

	return nil
}
