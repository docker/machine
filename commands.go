package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/docker/machine/drivers"
	_ "github.com/docker/machine/drivers/amazonec2"
	_ "github.com/docker/machine/drivers/azure"
	_ "github.com/docker/machine/drivers/digitalocean"
	_ "github.com/docker/machine/drivers/google"
	_ "github.com/docker/machine/drivers/none"
	_ "github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/state"
)

type hostListItem struct {
	Name       string
	Active     bool
	DriverName string
	State      state.State
	URL        string
}

type hostListItemByName []hostListItem

func (h hostListItemByName) Len() int {
	return len(h)
}

func (h hostListItemByName) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h hostListItemByName) Less(i, j int) bool {
	return strings.ToLower(h[i].Name) < strings.ToLower(h[j].Name)
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
					log.Fatalf("error getting active host: %v", err)
				}
				if host != nil {
					fmt.Println(host.Name)
				}
			} else if name != "" {
				host, err := store.Load(name)
				if err != nil {
					log.Fatalf("error loading host: %v", err)
				}

				if err := store.SetActive(host); err != nil {
					log.Fatalf("error setting active host: %v", err)
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
				log.Fatal(err)
			}

			if !keyExists {
				log.Fatalf("Identity authentication public key doesn't exist at %q. Create your public key by running the \"docker\" command.", drivers.PublicKeyPath())
			}

			store := NewStore()

			host, err := store.Create(name, driver, c)
			if err != nil {
				log.Fatal(err)
			}
			if err := store.SetActive(host); err != nil {
				log.Fatalf("error setting active host: %v", err)
			}

			log.Infof("%q has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity", name)
		},
	},
	{
		Name:  "inspect",
		Usage: "Inspect information about a machine",
		Action: func(c *cli.Context) {
			prettyJSON, err := json.MarshalIndent(getHost(c), "", "    ")
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(prettyJSON))
		},
	},
	{
		Name:  "ip",
		Usage: "Get the IP address of a machine",
		Action: func(c *cli.Context) {
			ip, err := getHost(c).Driver.GetIP()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(ip)
		},
	},
	{
		Name:  "kill",
		Usage: "Kill a machine",
		Action: func(c *cli.Context) {
			if err := getHost(c).Driver.Kill(); err != nil {
				log.Fatal(err)
			}
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
				log.Fatal(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

			if !quiet {
				fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL")
			}

			wg := sync.WaitGroup{}
			items := []hostListItem{}

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

						items = append(items, hostListItem{
							Name:       host.Name,
							Active:     isActive,
							DriverName: host.Driver.DriverName(),
							State:      currentState,
							URL:        url,
						})

						wg.Done()
					}()
				}
			}

			wg.Wait()

			sort.Sort(hostListItemByName(items))

			for _, item := range items {
				activeString := ""
				if item.Active {
					activeString = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					item.Name, activeString, item.DriverName, item.State, item.URL)
			}

			w.Flush()
		},
	},
	{
		Name:  "restart",
		Usage: "Restart a machine",
		Action: func(c *cli.Context) {
			if err := getHost(c).Driver.Restart(); err != nil {
				log.Fatal(err)
			}
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
				log.Fatal("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
			}
		},
	},
	{
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "command, c",
				Usage: "SSH Command",
				Value: "",
			},
		},
		Name:  "ssh",
		Usage: "Log into or run a command on a machine with SSH",
		Action: func(c *cli.Context) {
			name := c.Args().First()
			store := NewStore()

			if name == "" {
				host, err := store.GetActive()
				if err != nil {
					log.Fatalf("unable to get active host: %v", err)
				}

				name = host.Name
			}

			i := 1
			for i < len(os.Args) && os.Args[i-1] != name {
				i++
			}

			host, err := store.Load(name)
			if err != nil {
				log.Fatal(err)
			}

			var sshCmd *exec.Cmd
			if c.String("command") == "" {
				sshCmd, err = host.Driver.GetSSHCommand()
			} else {
				sshCmd, err = host.Driver.GetSSHCommand(c.String("command"))
			}
			if err != nil {
				log.Fatal(err)
			}

			sshCmd.Stdin = os.Stdin
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			if err := sshCmd.Run(); err != nil {
				log.Fatal(err)
			}
		},
	},
	{
		Name:  "start",
		Usage: "Start a machine",
		Action: func(c *cli.Context) {
			if err := getHost(c).Start(); err != nil {
				log.Fatal(err)
			}
		},
	},
	{
		Name:  "stop",
		Usage: "Stop a machine",
		Action: func(c *cli.Context) {
			if err := getHost(c).Stop(); err != nil {
				log.Fatal(err)
			}
		},
	},
	{
		Name:  "upgrade",
		Usage: "Upgrade a machine to the latest version of Docker",
		Action: func(c *cli.Context) {
			if err := getHost(c).Upgrade(); err != nil {
				log.Fatal(err)
			}
		},
	},
	{
		Name:  "url",
		Usage: "Get the URL of a machine",
		Action: func(c *cli.Context) {
			url, err := getHost(c).GetURL()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(url)
		},
	},
}

func getHost(c *cli.Context) *Host {
	name := c.Args().First()
	store := NewStore()

	if name == "" {
		host, err := store.GetActive()
		if err != nil {
			log.Fatalf("unable to get active host: %v", err)
		}
		return host
	}

	host, err := store.Load(name)
	if err != nil {
		log.Fatalf("unable to load host: %v", err)
	}
	return host
}
