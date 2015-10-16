package commands

import (
	"fmt"

	"github.com/docker/machine/cli"
)

func cmdActive(c *cli.Context) {
	if len(c.Args()) > 0 {
		fatal("Error: Too many arguments given.")
	}

	store := getStore(c)
	host, err := getActiveHost(store)
	if err != nil {
		fatalf("Error getting active host: %s", err)
	}

	if host != nil {
		fmt.Println(host.Name)
	}
}
