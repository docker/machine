package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
)

func cmdConfig(c *cli.Context) {
	if len(c.Args()) != 1 {
		log.Fatalln(ErrExpectedOneMachine)
	}
	cfg, err := getMachineConfig(c)
	if err != nil {
		log.Fatalln(err)
	}

	dockerHost, err := getHost(c).Driver.GetURL()
	if err != nil {
		log.Fatalln(err)
	}

	if c.Bool("swarm") {
		if !cfg.SwarmOptions.Master {
			log.Fatalf("%s is not a swarm master\n", cfg.machineName)
		}
		u, err := url.Parse(cfg.SwarmOptions.Host)
		if err != nil {
			log.Fatalln(err)
		}
		parts := strings.Split(u.Host, ":")
		swarmPort := parts[1]

		// get IP of machine to replace in case swarm host is 0.0.0.0
		mUrl, err := url.Parse(dockerHost)
		if err != nil {
			log.Fatalln(err)
		}
		mParts := strings.Split(mUrl.Host, ":")
		machineIp := mParts[0]

		dockerHost = fmt.Sprintf("tcp://%s:%s", machineIp, swarmPort)
	}

	log.Debugln(dockerHost)

	u, err := url.Parse(cfg.machineUrl)
	if err != nil {
		log.Fatalln(err)
	}

	if u.Scheme != "unix" {
		// validate cert and regenerate if needed
		valid, err := utils.ValidateCertificate(
			u.Host,
			cfg.caCertPath,
			cfg.serverCertPath,
			cfg.serverKeyPath,
		)
		if err != nil {
			log.Fatalln(err)
		}

		if !valid {
			log.Debugf("invalid certs detected; regenerating for %s\n", u.Host)

			if err := runActionWithContext("configureAuth", c); err != nil {
				log.Fatalln(err)
			}
		}
	}

	fmt.Printf("--tlsverify --tlscacert=%q --tlscert=%q --tlskey=%q -H=%s",
		cfg.caCertPath, cfg.clientCertPath, cfg.clientKeyPath, dockerHost)
}
