package commands

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdActive(c *cli.Context) {
	name := c.Args().First()

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

	if name == "" {
		host, err := mcn.GetActive()
		if err != nil {
			log.Fatalf("error getting active host: %v", err)
		}
		if host != nil {
			fmt.Println(host.Name)
		}
	} else if name != "" {
		host, err := mcn.Get(name)
		if err != nil {
			log.Fatalf("error loading host: %v", err)
		}

		if err := mcn.SetActive(host); err != nil {
			log.Fatalf("error setting active host: %v", err)
		}
	} else {
		cli.ShowCommandHelp(c, "active")
	}
}
