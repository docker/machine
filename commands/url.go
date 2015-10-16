package commands

import (
	"fmt"

	"github.com/docker/machine/cli"
)

func cmdUrl(c *cli.Context) {
	if len(c.Args()) != 1 {
		fatal(ErrExpectedOneMachine)
	}
	url, err := getFirstArgHost(c).GetURL()
	if err != nil {
		fatal(err)
	}

	fmt.Println(url)
}
