package commands

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/log"
)

func cmdActive(c *cli.Context) {
	if len(c.Args()) > 0 {
		log.Fatal("Error: Too many arguments given.")
	}

	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := newProvider(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	host, err := provider.GetActive()
	if err != nil {
		log.Fatalf("Error getting active host: %s", err)
	}

	if host != nil {
		fmt.Println(host.Name)
	}
}
