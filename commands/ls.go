package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/log"
)

// FilterOptions -
type FilterOptions struct {
	SwarmName  []string
	DriverName []string
	State      []string
}

func cmdLs(c *cli.Context) {
	quiet := c.Bool("quiet")
	filters, err := parseFilters(c.StringSlice("filter"))
	if err != nil {
		log.Fatal(err)
	}

	provider := getDefaultProvider(c)
	hostList, err := provider.List()
	if err != nil {
		log.Fatal(err)
	}

	hostList = filterHosts(hostList, filters)

	// Just print out the names if we're being quiet
	if quiet {
		for _, host := range hostList {
			fmt.Println(host.Name)
		}
		return
	}

	swarmMasters := make(map[string]string)
	swarmInfo := make(map[string]string)

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL\tSWARM")

	for _, host := range hostList {
		swarmOptions := host.HostOptions.SwarmOptions
		if swarmOptions.Master {
			swarmMasters[swarmOptions.Discovery] = host.Name
		}

		if swarmOptions.Discovery != "" {
			swarmInfo[host.Name] = swarmOptions.Discovery
		}
	}

	items := libmachine.GetHostListItems(hostList)

	sortHostListItemsByName(items)

	for _, item := range items {
		activeString := ""
		if item.Active {
			activeString = "*"
		}

		swarmInfo := ""

		if item.SwarmOptions.Discovery != "" {
			swarmInfo = swarmMasters[item.SwarmOptions.Discovery]
			if item.SwarmOptions.Master {
				swarmInfo = fmt.Sprintf("%s (master)", swarmInfo)
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Name, activeString, item.DriverName, item.State, item.URL, swarmInfo)
	}

	w.Flush()
}

func parseFilters(filters []string) (FilterOptions, error) {
	options := FilterOptions{}
	for _, f := range filters {
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			return options, errors.New("Unsupported filter syntax.")
		}
		key, value := kv[0], kv[1]

		switch key {
		case "swarm":
			options.SwarmName = append(options.SwarmName, value)
		case "driver":
			options.DriverName = append(options.DriverName, value)
		case "state":
			options.State = append(options.State, value)
		default:
			return options, fmt.Errorf("Unsupported filter key '%s'", key)
		}
	}
	return options, nil
}

func filterHosts(hosts []*libmachine.Host, filters FilterOptions) []*libmachine.Host {
	if len(filters.SwarmName) == 0 &&
		len(filters.DriverName) == 0 &&
		len(filters.State) == 0 {
		return hosts
	}

	filteredHosts := []*libmachine.Host{}
	swarmMasters := getSwarmMasters(hosts)

	for _, h := range hosts {
		if filterHost(h, filters, swarmMasters) {
			filteredHosts = append(filteredHosts, h)
		}
	}
	return filteredHosts
}

func getSwarmMasters(hosts []*libmachine.Host) map[string]string {
	swarmMasters := make(map[string]string)
	for _, h := range hosts {
		swarmOptions := h.HostOptions.SwarmOptions
		if swarmOptions != nil && swarmOptions.Master {
			swarmMasters[swarmOptions.Discovery] = h.Name
		}
	}
	return swarmMasters
}

func filterHost(host *libmachine.Host, filters FilterOptions, swarmMasters map[string]string) bool {
	swarmMatches := matchesSwarmName(host, filters.SwarmName, swarmMasters)
	driverMatches := matchesDriverName(host, filters.DriverName)
	stateMatches := matchesState(host, filters.State)

	return swarmMatches && driverMatches && stateMatches
}

func matchesSwarmName(host *libmachine.Host, swarmNames []string, swarmMasters map[string]string) bool {
	if len(swarmNames) == 0 {
		return true
	}
	for _, n := range swarmNames {
		if host.HostOptions.SwarmOptions != nil {
			if n == swarmMasters[host.HostOptions.SwarmOptions.Discovery] {
				return true
			}
		}
	}
	return false
}

func matchesDriverName(host *libmachine.Host, driverNames []string) bool {
	if len(driverNames) == 0 {
		return true
	}
	for _, n := range driverNames {
		if host.DriverName == n {
			return true
		}
	}
	return false
}

func matchesState(host *libmachine.Host, states []string) bool {
	if len(states) == 0 {
		return true
	}
	for _, n := range states {
		s, err := host.Driver.GetState()
		if err != nil {
			log.Warn(err)
		}
		if n == s.String() {
			return true
		}
	}
	return false
}
