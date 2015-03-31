package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func cmdRm(c *cli.Context) {
	if len(c.Args()) == 0 {
		cli.ShowCommandHelp(c, "rm")
		log.Fatal("You must specify a machine name")
	}

	force := c.Bool("force")

	isError := false

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

	for _, host := range c.Args() {
		if err := mcn.Remove(host, force); err != nil {
			log.Errorf("Error removing machine %s: %s", host, err)
			isError = true
		}
	}
	if isError {
		log.Fatal("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
	}
	log.Print("The machine was successfully removed.")
}
