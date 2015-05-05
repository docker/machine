package commands

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine"
)

func cmdLs(c *cli.Context) {
	quiet := c.Bool("quiet")

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	if !quiet {
		fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL\tSWARM")
	}

	mcn := getDefaultMcn(c)
	hostList, err := mcn.List()
	if err != nil {
		log.Fatal(err)
	}

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
		} else {
			fmt.Fprintf(w, "%s\n", host.Name)
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
