package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"

	"github.com/docker/machine/drivers"
	_ "github.com/docker/machine/drivers/azure"
	_ "github.com/docker/machine/drivers/cloudstack"
	_ "github.com/docker/machine/drivers/digitalocean"
	_ "github.com/docker/machine/drivers/none"
	_ "github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/state"
)

type DockerCli struct{}

func (cli *DockerCli) getMethod(args ...string) (func(...string) error, bool) {
	camelArgs := make([]string, len(args))
	for i, s := range args {
		if len(s) == 0 {
			return nil, false
		}
		camelArgs[i] = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	}
	methodName := "Cmd" + strings.Join(camelArgs, "")
	method := reflect.ValueOf(cli).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}

func (cli *DockerCli) Cmd(args ...string) error {
	if len(args) > 1 {
		method, exists := cli.getMethod(args[:2]...)
		if exists {
			return method(args[2:]...)
		}
	}
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Println("Error: Command not found:", args[0])
			return cli.CmdHelp()
		}
		return method(args[1:]...)
	}
	return cli.CmdHelp()
}

func (cli *DockerCli) Subcmd(name, signature, description string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		options := ""
		if flags.FlagCountUndeprecated() > 0 {
			options = "[OPTIONS] "
		}
		fmt.Fprintf(os.Stderr, "\nUsage: machine %s %s%s\n\n%s\n\n", name, options, signature, description)
		flags.PrintDefaults()
		os.Exit(2)
	}
	return flags
}

func (cli *DockerCli) CmdHelp(args ...string) error {
	if len(args) > 1 {
		method, exists := cli.getMethod(args[:2]...)
		if exists {
			method("--help")
			return nil
		}
	}
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Command not found: %s\n", args[0])
		} else {
			method("--help")
			return nil
		}
	}

	flag.Usage()

	return nil
}

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

func (cli *DockerCli) CmdLs(args ...string) error {
	cmd := cli.Subcmd("ls", "", "List machines")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display names")

	if err := cmd.Parse(args); err != nil {
		return err
	}

	store := NewStore()

	hostList, err := store.List()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	if !*quiet {
		fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL")
	}

	wg := sync.WaitGroup{}

	hostListItem := []HostListItem{}

	for _, host := range hostList {
		host := host
		if *quiet {
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

				hostListItem = append(hostListItem, HostListItem{
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

	sort.Sort(HostListItemByName(hostListItem))

	for _, hostState := range hostListItem {
		activeString := ""
		if hostState.Active {
			activeString = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			hostState.Name, activeString, hostState.DriverName, hostState.State, hostState.URL)
	}

	w.Flush()

	return nil
}

func (cli *DockerCli) CmdCreate(args ...string) error {
	cmd := cli.Subcmd("create", "NAME", "Create machines")

	driverDesc := fmt.Sprintf(
		"Driver to create machine with. Available drivers: %s",
		strings.Join(drivers.GetDriverNames(), ", "),
	)

	driver := cmd.String([]string{"d", "-driver"}, "none", driverDesc)

	createFlags := drivers.RegisterCreateFlags(cmd)

	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() != 1 {
		cmd.Usage()
		return nil
	}

	keyExists, err := drivers.PublicKeyExists()
	if err != nil {
		return err
	}
	if !keyExists {
		log.Fatalf("Identity auth public key does not exist at %s. Please run the docker client without any options to create it.", drivers.PublicKeyPath())
	}

	name := cmd.Arg(0)

	store := NewStore()

	driverCreateFlags, _ := createFlags[*driver]
	host, err := store.Create(name, *driver, driverCreateFlags)
	if err != nil {
		return err
	}
	if err := store.SetActive(host); err != nil {
		return err
	}
	log.Infof("%q has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity", name)
	return nil
}

func (cli *DockerCli) CmdStart(args ...string) error {
	cmd := cli.Subcmd("start", "NAME", "Start a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Start()
}

func (cli *DockerCli) CmdStop(args ...string) error {
	cmd := cli.Subcmd("stop", "NAME", "Stop a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Stop()
}

func (cli *DockerCli) CmdRm(args ...string) error {
	cmd := cli.Subcmd("rm", "NAME", "Remove a machine")
	force := cmd.Bool([]string{"f", "-force"}, false, "Remove local configuration even if machine cannot be removed")

	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	isError := false

	store := NewStore()
	for _, host := range cmd.Args() {
		host := host
		if err := store.Remove(host, *force); err != nil {
			log.Errorf("Error removing machine %s: %s", host, err)
			isError = true
		}
	}
	if isError {
		return fmt.Errorf("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
	}
	return nil
}

func (cli *DockerCli) CmdIp(args ...string) error {
	cmd := cli.Subcmd("ip", "NAME", "Get the IP address of a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() > 1 {
		cmd.Usage()
		return nil
	}

	var (
		err   error
		host  *Host
		store = NewStore()
	)

	if cmd.NArg() == 1 {
		host, err = store.Load(cmd.Arg(0))
		if err != nil {
			return err
		}
	} else {
		host, err = store.GetActive()
		if err != nil {
			return err
		}
		if host == nil {
			os.Exit(1)
		}
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return err
	}

	fmt.Println(ip)

	return nil
}

func (cli *DockerCli) CmdUrl(args ...string) error {
	cmd := cli.Subcmd("url", "NAME", "Get the URL of a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() > 1 {
		cmd.Usage()
		return nil
	}

	var (
		err   error
		host  *Host
		store = NewStore()
	)

	if cmd.NArg() == 1 {
		host, err = store.Load(cmd.Arg(0))
		if err != nil {
			return err
		}
	} else {
		host, err = store.GetActive()
		if err != nil {
			return err
		}
		if host == nil {
			os.Exit(1)
		}
	}

	url, err := host.GetURL()
	if err != nil {
		return err
	}

	fmt.Println(url)

	return nil
}

func (cli *DockerCli) CmdRestart(args ...string) error {
	cmd := cli.Subcmd("restart", "NAME", "Restart a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Driver.Restart()
}

func (cli *DockerCli) CmdKill(args ...string) error {
	cmd := cli.Subcmd("kill", "NAME", "Kill a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Driver.Kill()
}

func (cli *DockerCli) CmdSsh(args ...string) error {
	cmd := cli.Subcmd("ssh", "NAME [COMMAND ...]", "Log into or run a command on a machine with SSH")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	i := 1
	for i < len(os.Args) && os.Args[i-1] != cmd.Arg(0) {
		i++
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}

	sshCmd, err := host.Driver.GetSSHCommand(os.Args[i:]...)
	if err != nil {
		return err
	}
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func (cli *DockerCli) CmdActive(args ...string) error {
	cmd := cli.Subcmd("active", "[NAME]", "Get or set the active machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	store := NewStore()

	if cmd.NArg() == 0 {
		host, err := store.GetActive()
		if err != nil {
			return err
		}
		if host != nil {
			fmt.Println(host.Name)
		}
	} else if cmd.NArg() == 1 {
		host, err := store.Load(cmd.Arg(0))
		if err != nil {
			return err
		}
		if err := store.SetActive(host); err != nil {
			return err
		}
	} else {
		cmd.Usage()
	}

	return nil

}

func (cli *DockerCli) CmdInspect(args ...string) error {
	cmd := cli.Subcmd("inspect", "[NAME]", "Get detailed information about a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	if cmd.NArg() == 0 {
		cmd.Usage()
		return nil
	}

	store := NewStore()
	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}

	prettyJson, err := json.MarshalIndent(host, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(prettyJson))

	return nil
}

func (cli *DockerCli) CmdUpgrade(args ...string) error {
	cmd := cli.Subcmd("upgrade", "[NAME]", "Upgrade a machine to the latest version of Docker")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	if cmd.NArg() == 0 {
		cmd.Usage()
		return nil
	}

	store := NewStore()
	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}

	return host.Driver.Upgrade()
}
