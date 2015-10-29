package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/state"
)

var (
	errTooManyArguments = errors.New("Error: Too many arguments given")
)

func cmdActive(c CommandLine) error {
	if len(c.Args()) > 0 {
		return errTooManyArguments
	}

	store := getStore(c)

	host, err := getActiveHost(store)
	if err != nil {
		return fmt.Errorf("Error getting active host: %s", err)
	}

	if host != nil {
		fmt.Println(host.Name)
	}

	return nil
}

func getActiveHost(store persist.Store) (*host.Host, error) {
	hosts, err := listHosts(store)
	if err != nil {
		return nil, err
	}

	hostListItems := getHostListItems(hosts)

	for _, item := range hostListItems {
		if item.Active {
			return loadHost(store, item.Name)
		}
	}

	return nil, errors.New("Active host not found")
}

// IsActive provides a single function for determining if a host is active
// based on both the url and if the host is stopped.
func isActive(h *host.Host, currentState state.State, url string) (bool, error) {
	dockerHost := os.Getenv("DOCKER_HOST")

	// TODO: hard-coding the swarm port is a travesty...
	deSwarmedHost := strings.Replace(dockerHost, ":3376", ":2376", 1)
	if dockerHost == url || deSwarmedHost == url {
		return currentState == state.Running, nil
	}

	return false, nil
}
