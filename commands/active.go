package commands

import (
	"fmt"
	"time"

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

	timeout := time.Duration(c.Int("timeout")) * time.Second

	host, err := provider.GetActive(timeout)
	if err != nil {
		log.Fatalf("Error getting active host: %s", err)
	}

	if host != nil {
		fmt.Println(host.Name)
	}
}
