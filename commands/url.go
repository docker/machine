package commands

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
)

func cmdUrl(c *cli.Context) {
	if len(c.Args()) != 1 {
		log.Fatal(ErrExpectedOneMachine)
	}
	url, err := getFirstArgHost(c).GetURL()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(url)
}
