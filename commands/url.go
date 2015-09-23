package commands

import (
	"fmt"

	"github.com/docker/machine/libmachine/log"

	"github.com/codegangsta/cli"
)

func cmdUrl(c *cli.Context) {
	url, err := getFirstArgHost(c).GetURL()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(url)
}
