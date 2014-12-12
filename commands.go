package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/docker/machine/drivers"
	_ "github.com/docker/machine/drivers/azure"
	_ "github.com/docker/machine/drivers/digitalocean"
	_ "github.com/docker/machine/drivers/none"
	_ "github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/state"
)

type HostListItem struct {
	Name       string
	Active     bool
	DriverName string
	State      state.State
	URL        string
}

type HostListItemByName []HostListItem

func (h HostListItemByName) Len() int {
	return len(h)
}

func (h HostListItemByName) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h HostListItemByName) Less(i, j int) bool {
	return strings.ToLower(h[i].Name) < strings.ToLower(h[j].Name)
}

func BeforeHandler(c *cli.Context) error {
	// if c.Bool("debug") {
	// 	os.Setenv("DEBUG", "1")
	// 	initLogging(log.DebugLevel)
	// }

	return nil
}

var Commands = []cli.Command{
	{
		Name:  "active",
		Usage: "Get or set the active machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()
			store := NewStore()

			if name == "" {
				host, err := store.GetActive()
				if err != nil {
					log.Errorf("error finding active host")
				}
				if host != nil {
					fmt.Println(host.Name)
				}
			} else if name != "" {
				host, err := store.Load(name)
				if err != nil {
					log.Errorln(err)
					log.Errorf("error loading new active host")
					os.Exit(1)
				}

				if err := store.SetActive(host); err != nil {
					log.Errorf("error setting new active host")
					os.Exit(1)
				}
			} else {
				cli.ShowCommandHelp(c, "active")
			}
		},
	},
	{
		Flags: append(
			drivers.GetCreateFlags(),
			cli.StringFlag{
				Name: "driver, d",
				Usage: fmt.Sprintf(
					"Driver to create machine with. Available drivers: %s",
					strings.Join(drivers.GetDriverNames(), ", "),
				),
				Value: "none",
			},
		),
		Name:  "create",
		Usage: "Create a machine",
		Action: func(c *cli.Context) {
			driver := c.String("driver")
			name := c.Args().First()

			if name == "" {
				cli.ShowCommandHelp(c, "create")
				os.Exit(1)
			}

			keyExists, err := drivers.PublicKeyExists()
			if err != nil {
				log.Errorf("error")
				os.Exit(1)
			}

			if !keyExists {
				log.Errorf("error key doesn't exist")
				os.Exit(1)
			}

			store := NewStore()

			fmt.Printf("%#v", c.String("url"))

			host, err := store.Create(name, driver, c)
			if err != nil {
				log.Errorf("%s", err)
				os.Exit(1)
			}
			if err := store.SetActive(host); err != nil {
				log.Errorf("%s", err)
				os.Exit(1)
			}

			log.Infof("%q has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity", name)
		},
	},
	{
		Name:  "inspect",
		Usage: "Inspect information about a machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()

			if name == "" {
				cli.ShowCommandHelp(c, "inspect")
				os.Exit(1)
			}

			store := NewStore()
			host, err := store.Load(name)
			if err != nil {
				log.Errorf("error loading data")
				os.Exit(1)
			}

			prettyJson, err := json.MarshalIndent(host, "", "    ")
			if err != nil {
				log.Error("error with json")
				os.Exit(1)
			}

			fmt.Println(string(prettyJson))
		},
	},
	{
		Name:  "ip",
		Usage: "Get the IP address of a machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()

			if name == "" {
				cli.ShowCommandHelp(c, "ip")
				os.Exit(1)
			}

			var (
				err   error
				host  *Host
				store = NewStore()
			)

			if name != "" {
				host, err = store.Load(name)
				if err != nil {
					log.Errorf("error unable to load data")
					os.Exit(1)
				}
			} else {
				host, err = store.GetActive()
				if err != nil {
					log.Errorf("error")
					os.Exit(1)
				}
				if host == nil {
					os.Exit(1)
				}
			}

			ip, err := host.Driver.GetIP()
			if err != nil {
				log.Errorf("error unable to get IP")
				os.Exit(1)
			}

			fmt.Println(ip)
		},
	},
	{
		Name:  "kill",
		Usage: "Kill a machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()

			if name == "" {
				cli.ShowCommandHelp(c, "kill")
				os.Exit(1)
			}

			store := NewStore()

			host, err := store.Load(name)
			if err != nil {
				log.Errorf("error unable to load data")
				os.Exit(1)
			}

			host.Driver.Kill()
		},
	},
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet, q",
				Usage: "Enable quiet mode",
			},
		},
		Name:  "ls",
		Usage: "List machines",
		Action: func(c *cli.Context) {
			quiet := c.Bool("quiet")
			store := NewStore()

			hostList, err := store.List()
			if err != nil {
				log.Errorf("error unable to list hosts")
				os.Exit(1)
			}

			w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

			if !quiet {
				fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL")
			}

			wg := sync.WaitGroup{}

			for _, host := range hostList {
				host := host
				if quiet {
					fmt.Fprintf(w, "%s\n", host.Name)
				} else {
					wg.Add(1)
					go func() {
						currentState, err := host.Driver.GetState()
						if err != nil {
							log.Errorf("error getting state for host %s: %s", host.Name, err)
						}

						url, err := host.GetURL()
						if err != nil {
							if err == drivers.ErrHostIsNotRunning {
								url = ""
							} else {
								log.Errorf("error getting URL for host %s: %s", host.Name, err)
							}
						}

						isActive, err := store.IsActive(&host)
						if err != nil {
							log.Errorf("error determining whether host %q is active: %s",
								host.Name, err)
						}

						activeString := ""
						if isActive {
							activeString = "*"
						}

						fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
							host.Name, activeString, host.Driver.DriverName(), currentState, url)
						wg.Done()
					}()
				}
			}

			wg.Wait()
			w.Flush()
		},
	},
	{
		Name:  "restart",
		Usage: "Restart a machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()
			if name == "" {
				cli.ShowCommandHelp(c, "restart")
				os.Exit(1)
			}

			store := NewStore()

			host, err := store.Load(name)
			if err != nil {
				log.Errorf("error unable to load data")
				os.Exit(1)
			}

			host.Driver.Restart()
		},
	},
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Remove local configuration even if machine cannot be removed",
			},
		},
		Name:  "rm",
		Usage: "Remove a machine",
		Action: func(c *cli.Context) {
			if len(c.Args()) == 0 {
				cli.ShowCommandHelp(c, "rm")
				os.Exit(1)
			}

			force := c.Bool("force")

			isError := false

			store := NewStore()
			for _, host := range c.Args() {
				if err := store.Remove(host, force); err != nil {
					log.Errorf("Error removing machine %s: %s", host, err)
					isError = true
				}
			}
			if isError {
				log.Errorf("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
			}
		},
	},
	{
		Name:  "ssh",
		Usage: "Log into or run a command on a machine with SSH",
		Action: func(c *cli.Context) {
			name := c.Args().First()
			store := NewStore()

			if name == "" {
				host, err := store.GetActive()
				if err != nil {
					log.Errorf("error unable to get active host")
					os.Exit(1)
				}

				name = host.Name
			}

			i := 1
			for i < len(os.Args) && os.Args[i-1] != name {
				i++
			}

			host, err := store.Load(name)
			if err != nil {
				log.Errorf("%s", err)
				os.Exit(1)
			}

			sshCmd, err := host.Driver.GetSSHCommand(os.Args[i:]...)
			if err != nil {
				log.Errorf("%s", err)
				os.Exit(1)
			}

			sshCmd.Stdin = os.Stdin
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			if err := sshCmd.Run(); err != nil {
				log.Errorf("%s", err)
				os.Exit(1)
			}
		},
	},
	{
		Name:  "start",
		Usage: "Start a machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()
			store := NewStore()

			if name == "" {
				host, err := store.GetActive()
				if err != nil {
					log.Errorf("error unable to get active host")
					os.Exit(1)
				}

				name = host.Name
			}

			host, err := store.Load(name)
			if err != nil {
				log.Errorf("error unable to load data")
				os.Exit(1)
			}

			host.Start()
		},
	},
	{
		Name:  "stop",
		Usage: "Stop a machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()
			store := NewStore()

			if name == "" {
				host, err := store.GetActive()
				if err != nil {
					log.Errorf("error unable to get active host")
					os.Exit(1)
				}

				name = host.Name
			}

			host, err := store.Load(name)
			if err != nil {
				log.Errorf("error unable to load data")
				os.Exit(1)
			}

			host.Stop()
		},
	},
	{
		Name:  "upgrade",
		Usage: "Upgrade a machine to the latest version of Docker",
		Action: func(c *cli.Context) {
			name := c.Args().First()
			store := NewStore()

			if name == "" {
				host, err := store.GetActive()
				if err != nil {
					log.Errorf("error unable to get active host")
					os.Exit(1)
				}

				name = host.Name
			}

			host, err := store.Load(name)
			if err != nil {
				log.Errorf("error unable to load host")
				os.Exit(1)
			}

			host.Driver.Upgrade()
		},
	},
	{
		Name:  "url",
		Usage: "Get the URL of a machine",
		Action: func(c *cli.Context) {
			name := c.Args().First()

			if name == "" {
				cli.ShowCommandHelp(c, "url")
				os.Exit(1)
			}

			var (
				err   error
				host  *Host
				store = NewStore()
			)

			if name != "" {
				host, err = store.Load(name)
				if err != nil {
					log.Errorf("error unable to load data")
					os.Exit(1)
				}
			} else {
				host, err = store.GetActive()
				if err != nil {
					log.Errorf("error unable to get active host")
					os.Exit(1)
				}
				if host == nil {
					os.Exit(1)
				}
			}

			url, err := host.GetURL()
			if err != nil {
				log.Errorf("error unable to get url for host")
				os.Exit(1)
			}

			fmt.Println(url)
		},
	},
}
