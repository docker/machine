package commands

import (
	"errors"
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/persist"
)

var (
	errNoActiveHost = errors.New("No active host found")
)

func cmdActive(c CommandLine, api libmachine.API) error {
	if len(c.Args()) > 0 {
		return ErrTooManyArguments
	}

	hosts, hostsInError, err := persist.LoadAllHosts(api)
	if err != nil {
		return fmt.Errorf("Error getting active host: %s", err)
	}

	items := getHostListItems(hosts, hostsInError)

	for _, item := range items {
		if item.ActiveHost {
			fmt.Println(item.Name)
			return nil
		}
	}

	return errNoActiveHost
}
