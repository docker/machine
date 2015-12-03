package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine"
)

func cmdURL(c CommandLine, api libmachine.API) error {
	if len(c.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := api.Load(c.Args().First())
	if err != nil {
		return err
	}

	url, err := host.URL()
	if err != nil {
		return err
	}

	fmt.Println(url)

	return nil
}
