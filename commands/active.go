package commands

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
)

func cmdActive(c *cli.Context) {
	if len(c.Args()) > 0 {
		log.Fatal("Error: Too many arguments given.")
	}

	store := getStore(c)
	host, err := getActiveHost(store)
	if err != nil {
		log.Fatalf("Error getting active host: %s", err)
	}

	if host != nil {
		fmt.Println(host.Name)
	}
}
