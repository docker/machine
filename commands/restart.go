package commands

import (
	"os"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
)

func cmdRestart(c *cli.Context) {
	if err := runActionWithContext("restart", c); err != nil {
		log.Fatal(err)
	}

	warnIfActiveIPChanged(c)
}

func warnIfActiveIPChanged(c *cli.Context) {
	if activeInList(c.Args()) {
		activeURL := os.Getenv("DOCKER_HOST")
		log.Debugf("checking to see if active host's URL has changed from %s", activeURL)

		h := getActiveHost(c)
		u, err := h.GetURL()
		if err != nil {
			log.Fatal(err)
		}
		if activeURL != u {
			// TODO: hardcoded port here is a giant kludge...
			var swarmOpt string
			if strings.Contains(activeURL, ":3376") {
				swarmOpt = "--swarm "
			}
			log.Warnf("Active machine was restarted, and has a new IP address.\nRun 'eval \"$(docker-machine env %s%s)\"' again.", swarmOpt, h.Name)
		}
	}
}

func activeInList(machines []string) bool {
	active := os.Getenv("DOCKER_MACHINE_NAME")
	return active != "" && inArray(active, machines)
}

func getActiveHost(c *cli.Context) *libmachine.Host {
	certInfo := getCertPathInfo(c)
	store, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}
	active, err := store.GetActive()
	if err != nil {
		log.Fatal(err)
	}
	return active
}

func inArray(s string, a []string) bool {
	for _, cur := range a {
		if cur == s {
			return true
		}
	}
	return false
}
