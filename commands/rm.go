package commands

import (
	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
)

func cmdRm(c *cli.Context) {
	if len(c.Args()) == 0 {
		cli.ShowCommandHelp(c, "rm")
		log.Fatal("You must specify a machine name")
	}

	force := c.Bool("force")

	isError := false

	store := getStore(c)

	for _, hostName := range c.Args() {
		if err := store.Remove(hostName, force); err != nil {
			log.Errorf("Error removing machine %s: %s", hostName, err)
			isError = true
		} else {
			log.Infof("Successfully removed %s", hostName)
		}
	}
	if isError {
		log.Fatal("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
	}
}
