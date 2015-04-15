package commands

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/utils"
)

func cmdEnv(c *cli.Context) {
	userShell := filepath.Base(os.Getenv("SHELL"))
	if c.Bool("unset") {
		switch userShell {
		case "fish":
			fmt.Printf("set -e DOCKER_TLS_VERIFY;\nset -e DOCKER_CERT_PATH;\nset -e DOCKER_HOST;\n")
		default:
			fmt.Println("unset DOCKER_TLS_VERIFY DOCKER_CERT_PATH DOCKER_HOST")
		}
		return
	}

	cfg, err := getMachineConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.machineUrl == "" {
		log.Fatalf("%s is not running. Please start this with docker-machine start %s", cfg.machineName, cfg.machineName)
	}

	dockerHost := cfg.machineUrl
	if c.Bool("swarm") {
		if !cfg.SwarmOptions.Master {
			log.Fatalf("%s is not a swarm master", cfg.machineName)
		}
		u, err := url.Parse(cfg.SwarmOptions.Host)
		if err != nil {
			log.Fatal(err)
		}
		parts := strings.Split(u.Host, ":")
		swarmPort := parts[1]

		// get IP of machine to replace in case swarm host is 0.0.0.0
		mUrl, err := url.Parse(cfg.machineUrl)
		if err != nil {
			log.Fatal(err)
		}
		mParts := strings.Split(mUrl.Host, ":")
		machineIp := mParts[0]

		dockerHost = fmt.Sprintf("tcp://%s:%s", machineIp, swarmPort)
	}

	u, err := url.Parse(cfg.machineUrl)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}

		if !valid {
			log.Debugf("invalid certs detected; regenerating for %s", u.Host)

			if err := runActionWithContext("configureAuth", c); err != nil {
				log.Fatal(err)
			}
		}
	}

	usageHint := generateUsageHint(c.Args().First(), userShell)

	switch userShell {
	case "fish":
		fmt.Printf("set -x DOCKER_TLS_VERIFY 1;\nset -x DOCKER_CERT_PATH %q;\nset -x DOCKER_HOST %s;\n\n%s\n",
			cfg.machineDir, dockerHost, usageHint)
	default:
		fmt.Printf("export DOCKER_TLS_VERIFY=1\nexport DOCKER_CERT_PATH=%q\nexport DOCKER_HOST=%s\n\n%s\n",
			cfg.machineDir, dockerHost, usageHint)
	}
}

func generateUsageHint(machineName string, userShell string) string {
	cmd := ""
	switch userShell {
	case "fish":
		if machineName != "" {
			cmd = fmt.Sprintf("eval (docker-machine env %s)", machineName)
		} else {
			cmd = "eval (docker-machine env)"
		}
	default:
		if machineName != "" {
			cmd = fmt.Sprintf("eval \"$(docker-machine env %s)\"", machineName)
		} else {
			cmd = "eval \"$(docker-machine env)\""
		}
	}

	return fmt.Sprintf("# Run this command to configure your shell: %s\n", cmd)
}
