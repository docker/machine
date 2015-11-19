package commands

import (
	"errors"
	"fmt"

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/persist"
)

var (
	errTooManyArguments = errors.New("Error: Too many arguments given")
)

func cmdActive(cli CommandLine, store persist.Store) error {
	if len(cli.Args()) > 0 {
		return errTooManyArguments
	}

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
