package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine/persist"
)

func cmdURL(c CommandLine, store persist.Store) error {
	if len(c.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := store.Load(c.Args().First())
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
