package commands

import (
	"fmt"

	"github.com/docker/machine/cli"
)

func cmdURL(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return ErrExpectedOneMachine
	}

	host, err := getFirstArgHost(c)
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
