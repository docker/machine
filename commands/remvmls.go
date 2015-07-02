package commands

import (
	"github.com/codegangsta/cli"
	"github.com/docker/machine/log"
)

func cmdRemVMLs(c *cli.Context) {
	var (
		err error
	)

	driver := c.String("driver")

	// TODO: Not really a fan of "none" as the default driver...
	if driver != "none" {
		c.App.Commands, err = trimDriverFlags(driver, c.App.Commands)
		if err != nil {
			log.Fatal(err)
		}
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

	mcn, err := newMcn(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	err = mcn.RemVMLs(driver, c)
	if err != nil {
		log.Errorf("Error listing instance: %s", err)
	}
}
