package commands

import (
	"fmt"
)

func cmdURL(c CommandLine) error {
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
