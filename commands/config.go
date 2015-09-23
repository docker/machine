package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"
)

func cmdConfig(c *cli.Context) {
	if len(c.Args()) != 1 {
		log.Fatal(ErrExpectedOneMachine)
	}

	h := getFirstArgHost(c)

	dockerHost, authOptions, err := runConnectionBoilerplate(h, c)
	if err != nil {
		log.Fatalf("Error running connection boilerplate: %s", err)
	}

	log.Debug(dockerHost)

	fmt.Printf("--tlsverify --tlscacert=%q --tlscert=%q --tlskey=%q -H=%s",
		authOptions.CaCertPath, authOptions.ClientCertPath, authOptions.ClientKeyPath, dockerHost)
}

func runConnectionBoilerplate(h *host.Host, c *cli.Context) (string, *auth.AuthOptions, error) {
	hostState, err := h.Driver.GetState()
	if err != nil {
		// TODO: This is a common operation and should have a commonly
		// defined error.
		return "", &auth.AuthOptions{}, fmt.Errorf("Error trying to get host state: %s", err)
	}
	if hostState != state.Running {
		return "", &auth.AuthOptions{}, fmt.Errorf("%s is not running. Please start it in order to use the connection settings.", h.Name)
	}

	dockerHost, err := h.Driver.GetURL()
	if err != nil {
		return "", &auth.AuthOptions{}, fmt.Errorf("Error getting driver URL: %s", err)
	}

	if c.Bool("swarm") {
		var err error
		dockerHost, err = parseSwarm(dockerHost, h)
		if err != nil {
			return "", &auth.AuthOptions{}, fmt.Errorf("Error parsing swarm: %s", err)
		}
	}

	u, err := url.Parse(dockerHost)
	if err != nil {
		return "", &auth.AuthOptions{}, fmt.Errorf("Error parsing URL: %s", err)
	}

	authOptions := h.HostOptions.AuthOptions

	if err := checkCert(u.Host, authOptions, c); err != nil {
		return "", &auth.AuthOptions{}, fmt.Errorf("Error checking and/or regenerating the certs: %s", err)
	}

	return dockerHost, authOptions, nil
}

func checkCert(hostUrl string, authOptions *auth.AuthOptions, c *cli.Context) error {
	valid, err := cert.ValidateCertificate(
		hostUrl,
		authOptions.CaCertPath,
		authOptions.ServerCertPath,
		authOptions.ServerKeyPath,
	)
	if err != nil {
		return fmt.Errorf("Error attempting to validate the certficate: %s", err)
	}

	if !valid {
		log.Errorf("Invalid certs detected; regenerating for %s", hostUrl)

		if err := runActionWithContext("configureAuth", c); err != nil {
			return fmt.Errorf("Error attempting to regenerate the certs: %s", err)
		}
	}

	return nil
}

// TODO: This could use a unit test.
func parseSwarm(hostUrl string, h *host.Host) (string, error) {
	swarmOptions := h.HostOptions.SwarmOptions

	if !swarmOptions.Master {
		return "", fmt.Errorf("Error: %s is not a swarm master.  The --swarm flag is intended for use with swarm masters.", h.Name)
	}

	u, err := url.Parse(swarmOptions.Host)
	if err != nil {
		return "", fmt.Errorf("There was an error parsing the url: %s", err)
	}
	parts := strings.Split(u.Host, ":")
	swarmPort := parts[1]

	// get IP of machine to replace in case swarm host is 0.0.0.0
	mUrl, err := url.Parse(hostUrl)
	if err != nil {
		return "", fmt.Errorf("There was an error parsing the url: %s", err)
	}

	mParts := strings.Split(mUrl.Host, ":")
	machineIp := mParts[0]

	hostUrl = fmt.Sprintf("tcp://%s:%s", machineIp, swarmPort)

	return hostUrl, nil
}
