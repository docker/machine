package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine/persist"
)

func cmdURL(cli CommandLine, store persist.Store) error {
	if len(cli.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := loadHost(store, cli.Args().First())
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
