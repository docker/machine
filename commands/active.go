package commands

import (
	"github.com/codegangsta/cli"
	"github.com/docker/machine/log"
)

func cmdActive(c *cli.Context) {
	if len(c.Args()) > 0 {
		log.Fatalln("Error: Too many arguments given.")
	}

	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatalln(err)
	}

	mcn, err := newMcn(defaultStore)
	if err != nil {
		log.Fatalln(err)
	}

	host, err := mcn.GetActive()
	if err != nil {
		log.Fatalf("Error getting active host: %s\n", err)
	}

	if host != nil {
		log.Println(host.Name)
	}
}
