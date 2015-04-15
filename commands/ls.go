package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
)

func cmdLs(c *cli.Context) {
	quiet := c.Bool("quiet")

	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	mcn, err := newMcn(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	hostList, err := mcn.List()
	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	if !quiet {
		fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL\tSWARM")
	}

	items := []hostListItem{}
	hostListItems := make(chan hostListItem)

	swarmMasters := make(map[string]string)
	swarmInfo := make(map[string]string)

	for _, host := range hostList {
		swarmOptions := host.HostOptions.SwarmOptions
		if !quiet {
			if swarmOptions.Master {
				swarmMasters[swarmOptions.Discovery] = host.Name
			}

			if swarmOptions.Discovery != "" {
				swarmInfo[host.Name] = swarmOptions.Discovery
			}

			go getHostState(*host, defaultStore, hostListItems)
		} else {
			fmt.Fprintf(w, "%s\n", host.Name)
		}
	}

	if !quiet {
		for i := 0; i < len(hostList); i++ {
			items = append(items, <-hostListItems)
		}
	}

	close(hostListItems)

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
