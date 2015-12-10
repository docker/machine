package commands

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/skarademir/naturalsort"
)

const lsDefaultTimeout = 10

var (
	stateTimeoutDuration = lsDefaultTimeout * time.Second
)

// FilterOptions -
type FilterOptions struct {
	SwarmName  []string
	DriverName []string
	State      []string
	Name       []string
	Labels     []string
}

type HostListItem struct {
	Name          string
	Active        bool
	DriverName    string
	State         state.State
	URL           string
	SwarmOptions  *swarm.Options
	EngineOptions *engine.Options
	Error         string
	DockerVersion string
}

func cmdLs(c CommandLine, api libmachine.API) error {
	stateTimeoutDuration = time.Duration(c.Int("timeout")) * time.Second
	log.Debugf("ls timeout set to %s", stateTimeoutDuration)

	quiet := c.Bool("quiet")
	filters, err := parseFilters(c.StringSlice("filter"))
	if err != nil {
		return err
	}

	hostList, hostInError, err := persist.LoadAllHosts(api)
	if err != nil {
		return err
	}

	hostList = filterHosts(hostList, filters)

	// Just print out the names if we're being quiet
	if quiet {
		for _, host := range hostList {
			fmt.Println(host.Name)
		}
		return nil
	}

	swarmMasters := make(map[string]string)
	swarmInfo := make(map[string]string)

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL\tSWARM\tDOCKER\tERRORS")

	for _, host := range hostList {
		swarmOptions := host.HostOptions.SwarmOptions
		if swarmOptions.Master {
			swarmMasters[swarmOptions.Discovery] = host.Name
		}

		if swarmOptions.Discovery != "" {
			swarmInfo[host.Name] = swarmOptions.Discovery
		}
	}

	items := getHostListItems(hostList, hostInError)

	for _, item := range items {
		activeString := "-"
		if item.Active {
			activeString = "*"
		}

		swarmInfo := ""

		if item.SwarmOptions != nil && item.SwarmOptions.Discovery != "" {
			swarmInfo = swarmMasters[item.SwarmOptions.Discovery]
			if item.SwarmOptions.Master {
				swarmInfo = fmt.Sprintf("%s (master)", swarmInfo)
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Name, activeString, item.DriverName, item.State, item.URL, swarmInfo, item.DockerVersion, item.Error)
	}

	w.Flush()

	return nil
}

func parseFilters(filters []string) (FilterOptions, error) {
	options := FilterOptions{}
	for _, f := range filters {
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			return options, errors.New("Unsupported filter syntax.")
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		switch key {
		case "swarm":
			options.SwarmName = append(options.SwarmName, value)
		case "driver":
			options.DriverName = append(options.DriverName, value)
		case "state":
			options.State = append(options.State, value)
		case "name":
			options.Name = append(options.Name, value)
		case "label":
			options.Labels = append(options.Labels, value)
		default:
			return options, fmt.Errorf("Unsupported filter key '%s'", key)
		}
	}
	return options, nil
}

func filterHosts(hosts []*host.Host, filters FilterOptions) []*host.Host {
	if len(filters.SwarmName) == 0 &&
		len(filters.DriverName) == 0 &&
		len(filters.State) == 0 &&
		len(filters.Name) == 0 &&
		len(filters.Labels) == 0 {
		return hosts
	}

	filteredHosts := []*host.Host{}
	swarmMasters := getSwarmMasters(hosts)

	for _, h := range hosts {
		if filterHost(h, filters, swarmMasters) {
			filteredHosts = append(filteredHosts, h)
		}
	}
	return filteredHosts
}

func getSwarmMasters(hosts []*host.Host) map[string]string {
	swarmMasters := make(map[string]string)
	for _, h := range hosts {
		if h.HostOptions != nil {
			swarmOptions := h.HostOptions.SwarmOptions
			if swarmOptions != nil && swarmOptions.Master {
				swarmMasters[swarmOptions.Discovery] = h.Name
			}
		}
	}
	return swarmMasters
}

func filterHost(host *host.Host, filters FilterOptions, swarmMasters map[string]string) bool {
	swarmMatches := matchesSwarmName(host, filters.SwarmName, swarmMasters)
	driverMatches := matchesDriverName(host, filters.DriverName)
	stateMatches := matchesState(host, filters.State)
	nameMatches := matchesName(host, filters.Name)
	labelMatches := matchesLabel(host, filters.Labels)

	return swarmMatches && driverMatches && stateMatches && nameMatches && labelMatches
}

func matchesSwarmName(host *host.Host, swarmNames []string, swarmMasters map[string]string) bool {
	if len(swarmNames) == 0 {
		return true
	}
	for _, n := range swarmNames {
		if host.HostOptions != nil && host.HostOptions.SwarmOptions != nil {
			if strings.EqualFold(n, swarmMasters[host.HostOptions.SwarmOptions.Discovery]) {
				return true
			}
		}
	}
	return false
}

func matchesDriverName(host *host.Host, driverNames []string) bool {
	if len(driverNames) == 0 {
		return true
	}
	for _, n := range driverNames {
		if strings.EqualFold(host.DriverName, n) {
			return true
		}
	}
	return false
}

func matchesState(host *host.Host, states []string) bool {
	if len(states) == 0 {
		return true
	}
	for _, n := range states {
		s, err := host.Driver.GetState()
		if err != nil {
			log.Warn(err)
		}
		if strings.EqualFold(n, s.String()) {
			return true
		}
	}
	return false
}

func matchesName(host *host.Host, names []string) bool {
	if len(names) == 0 {
		return true
	}
	for _, n := range names {
		r, err := regexp.Compile(n)
		if err != nil {
			// TODO: remove that call to Fatal
			log.Fatal(err)
		}
		if r.MatchString(host.Driver.GetMachineName()) {
			return true
		}
	}
	return false
}

func matchesLabel(host *host.Host, labels []string) bool {
	if len(labels) == 0 {
		return true
	}

	var englabels = make(map[string]string, len(host.HostOptions.EngineOptions.Labels))

	if host.HostOptions != nil && host.HostOptions.EngineOptions.Labels != nil {
		for _, s := range host.HostOptions.EngineOptions.Labels {
			kv := strings.SplitN(s, "=", 2)
			englabels[kv[0]] = kv[1]
		}
	}

	for _, l := range labels {
		kv := strings.SplitN(l, "=", 2)
		if val, exists := englabels[kv[0]]; exists && strings.EqualFold(val, kv[1]) {
			return true
		}
	}
	return false
}

func attemptGetHostState(h *host.Host, stateQueryChan chan<- HostListItem) {
	url := ""
	hostError := ""
	dockerVersion := "Unknown"

	currentState, err := h.Driver.GetState()
	if err == nil {
		url, err = h.URL()
	}
	if err == nil {
		dockerVersion, err = h.DockerVersion()
		if err != nil {
			dockerVersion = "Unknown"
		} else {
			dockerVersion = fmt.Sprintf("v%s", dockerVersion)
		}
	}

	if err != nil {
		hostError = err.Error()
	}
	if hostError == drivers.ErrHostIsNotRunning.Error() {
		hostError = ""
	}

	var swarmOptions *swarm.Options
	var engineOptions *engine.Options
	if h.HostOptions != nil {
		swarmOptions = h.HostOptions.SwarmOptions
		engineOptions = h.HostOptions.EngineOptions
	}

	stateQueryChan <- HostListItem{
		Name:          h.Name,
		Active:        isActive(currentState, url),
		DriverName:    h.Driver.DriverName(),
		State:         currentState,
		URL:           url,
		SwarmOptions:  swarmOptions,
		EngineOptions: engineOptions,
		DockerVersion: dockerVersion,
		Error:         hostError,
	}
}

func getHostState(h *host.Host, hostListItemsChan chan<- HostListItem) {
	// This channel is used to communicate the properties we are querying
	// about the host in the case of a successful read.
	stateQueryChan := make(chan HostListItem)

	go attemptGetHostState(h, stateQueryChan)

	select {
	// If we get back useful information, great.  Forward it straight to
	// the original parent channel.
	case hli := <-stateQueryChan:
		hostListItemsChan <- hli

	// Otherwise, give up after a predetermined duration.
	case <-time.After(stateTimeoutDuration):
		hostListItemsChan <- HostListItem{
			Name:       h.Name,
			DriverName: h.Driver.DriverName(),
			State:      state.Timeout,
		}
	}
}

func getHostListItems(hostList []*host.Host, hostsInError map[string]error) []HostListItem {
	hostListItems := []HostListItem{}
	hostListItemsChan := make(chan HostListItem)

	for _, h := range hostList {
		go getHostState(h, hostListItemsChan)
	}

	for range hostList {
		hostListItems = append(hostListItems, <-hostListItemsChan)
	}

	close(hostListItemsChan)

	for name, err := range hostsInError {
		itemInError := HostListItem{}
		itemInError.Name = name
		itemInError.DriverName = "not found"
		itemInError.State = state.Error
		itemInError.Error = err.Error()
		hostListItems = append(hostListItems, itemInError)
	}

	sortHostListItemsByName(hostListItems)
	return hostListItems
}

func sortHostListItemsByName(items []HostListItem) {
	m := make(map[string]HostListItem, len(items))
	s := make([]string, len(items))
	for i, v := range items {
		name := strings.ToLower(v.Name)
		m[name] = v
		s[i] = name
	}
	sort.Sort(naturalsort.NaturalSort(s))
	for i, v := range s {
		items[i] = m[v]
	}
}

// IsActive provides a single function for determining if a host is active
// based on both the url and if the host is stopped.
func isActive(currentState state.State, url string) bool {
	dockerHost := os.Getenv("DOCKER_HOST")

	// TODO: hard-coding the swarm port is a travesty...
	deSwarmedHost := strings.Replace(dockerHost, ":3376", ":2376", 1)
	if dockerHost == url || deSwarmedHost == url {
		return currentState == state.Running
	}

	return false
}
