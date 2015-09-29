package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	_ "github.com/docker/machine/drivers/amazonec2"
	_ "github.com/docker/machine/drivers/azure"
	_ "github.com/docker/machine/drivers/digitalocean"
	_ "github.com/docker/machine/drivers/exoscale"
	_ "github.com/docker/machine/drivers/generic"
	_ "github.com/docker/machine/drivers/google"
	_ "github.com/docker/machine/drivers/hyperv"
	_ "github.com/docker/machine/drivers/none"
	_ "github.com/docker/machine/drivers/openstack"
	_ "github.com/docker/machine/drivers/rackspace"
	_ "github.com/docker/machine/drivers/softlayer"
	_ "github.com/docker/machine/drivers/virtualbox"
	_ "github.com/docker/machine/drivers/vmwarefusion"
	_ "github.com/docker/machine/drivers/vmwarevcloudair"
	_ "github.com/docker/machine/drivers/vmwarevsphere"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

var (
	ErrUnknownShell       = errors.New("Error: Unknown shell")
	ErrNoMachineSpecified = errors.New("Error: Expected to get one or more machine names as arguments.")
	ErrExpectedOneMachine = errors.New("Error: Expected one machine name as an argument.")
)

func confirmInput(msg string) bool {
	fmt.Printf("%s (y/n): ", msg)
	var resp string
	_, err := fmt.Scanln(&resp)

	if err != nil {
		log.Fatal(err)

	}

	if strings.Index(strings.ToLower(resp), "y") == 0 {
		return true

	}

	return false
}

func getStore(c *cli.Context) persist.Store {
	certInfo := getCertPathInfoFromContext(c)
	return &persist.Filestore{
		Path:             c.GlobalString("storage-path"),
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	}
}

func getFirstArgHost(c *cli.Context) *host.Host {
	store := getStore(c)
	hostName := c.Args().First()
	h, err := store.Load(hostName)
	if err != nil {
		// I guess I feel OK with bailing here since if we can't get
		// the host reliably we're definitely not going to be able to
		// do anything else interesting, but also this premature exit
		// feels wrong to me.  Let's revisit it later.
		log.Fatalf("Error trying to get host %q: %s", hostName, err)
	}
	return h
}

func getHostsFromContext(c *cli.Context) ([]*host.Host, error) {
	store := getStore(c)
	hosts := []*host.Host{}

	for _, hostName := range c.Args() {
		h, err := store.Load(hostName)
		if err != nil {
			return nil, fmt.Errorf("Could not load host %q: %s", hostName, err)
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
}

var sharedCreateFlags = []cli.Flag{
	cli.StringFlag{
		Name: "driver, d",
		Usage: fmt.Sprintf(
			"Driver to create machine with. Available drivers: %s",
			strings.Join(drivers.GetDriverNames(), ", "),
		),
		Value: "none",
	},
	cli.StringFlag{
		Name:   "engine-install-url",
		Usage:  "Custom URL to use for engine installation",
		Value:  "https://get.docker.com",
		EnvVar: "MACHINE_DOCKER_INSTALL_URL",
	},
	cli.StringSliceFlag{
		Name:  "engine-opt",
		Usage: "Specify arbitrary flags to include with the created engine in the form flag=value",
		Value: &cli.StringSlice{},
	},
	cli.StringSliceFlag{
		Name:  "engine-insecure-registry",
		Usage: "Specify insecure registries to allow with the created engine",
		Value: &cli.StringSlice{},
	},
	cli.StringSliceFlag{
		Name:  "engine-registry-mirror",
		Usage: "Specify registry mirrors to use",
		Value: &cli.StringSlice{},
	},
	cli.StringSliceFlag{
		Name:  "engine-label",
		Usage: "Specify labels for the created engine",
		Value: &cli.StringSlice{},
	},
	cli.StringFlag{
		Name:  "engine-storage-driver",
		Usage: "Specify a storage driver to use with the engine",
	},
	cli.StringSliceFlag{
		Name:  "engine-env",
		Usage: "Specify environment variables to set in the engine",
		Value: &cli.StringSlice{},
	},
	cli.BoolFlag{
		Name:  "swarm",
		Usage: "Configure Machine with Swarm",
	},
	cli.StringFlag{
		Name:   "swarm-image",
		Usage:  "Specify Docker image to use for Swarm",
		Value:  "swarm:latest",
		EnvVar: "MACHINE_SWARM_IMAGE",
	},
	cli.BoolFlag{
		Name:  "swarm-master",
		Usage: "Configure Machine to be a Swarm master",
	},
	cli.StringFlag{
		Name:  "swarm-discovery",
		Usage: "Discovery service to use with Swarm",
		Value: "",
	},
	cli.StringFlag{
		Name:  "swarm-strategy",
		Usage: "Define a default scheduling strategy for Swarm",
		Value: "spread",
	},
	cli.StringSliceFlag{
		Name:  "swarm-opt",
		Usage: "Define arbitrary flags for swarm",
		Value: &cli.StringSlice{},
	},
	cli.StringFlag{
		Name:  "swarm-host",
		Usage: "ip/socket to listen on for Swarm master",
		Value: "tcp://0.0.0.0:3376",
	},
	cli.StringFlag{
		Name:  "swarm-addr",
		Usage: "addr to advertise for Swarm (default: detect and use the machine IP)",
		Value: "",
	},
}

var Commands = []cli.Command{
	{
		Name:   "active",
		Usage:  "Print which machine is active",
		Action: cmdActive,
	},
	{
		Name:        "config",
		Usage:       "Print the connection config for machine",
		Description: "Argument is a machine name.",
		Action:      cmdConfig,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "swarm",
				Usage: "Display the Swarm config instead of the Docker daemon",
			},
		},
	},
	{
		Flags: append(
			drivers.GetCreateFlags(),
			sharedCreateFlags...,
		),
		Name:   "create",
		Usage:  "Create a machine",
		Action: cmdCreate,
	},
	{
		Name:        "env",
		Usage:       "Display the commands to set up the environment for the Docker client",
		Description: "Argument is a machine name.",
		Action:      cmdEnv,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "swarm",
				Usage: "Display the Swarm config instead of the Docker daemon",
			},
			cli.StringFlag{
				Name:  "shell",
				Usage: "Force environment to be configured for specified shell",
			},
			cli.BoolFlag{
				Name:  "unset, u",
				Usage: "Unset variables instead of setting them",
			},
			cli.BoolFlag{
				Name:  "no-proxy",
				Usage: "Add machine IP to NO_PROXY environment variable",
			},
		},
	},
	{
		Name:        "inspect",
		Usage:       "Inspect information about a machine",
		Description: "Argument is a machine name.",
		Action:      cmdInspect,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format, f",
				Usage: "Format the output using the given go template.",
				Value: "",
			},
		},
	},
	{
		Name:        "ip",
		Usage:       "Get the IP address of a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdIp,
	},
	{
		Name:        "kill",
		Usage:       "Kill a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdKill,
	},
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet, q",
				Usage: "Enable quiet mode",
			},
			cli.StringSliceFlag{
				Name:  "filter",
				Usage: "Filter output based on conditions provided",
				Value: &cli.StringSlice{},
			},
		},
		Name:   "ls",
		Usage:  "List machines",
		Action: cmdLs,
	},
	{
		Name:        "regenerate-certs",
		Usage:       "Regenerate TLS Certificates for a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdRegenerateCerts,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Force rebuild and do not prompt",
			},
		},
	},
	{
		Name:        "restart",
		Usage:       "Restart a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdRestart,
	},
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Remove local configuration even if machine cannot be removed",
			},
		},
		Name:        "rm",
		Usage:       "Remove a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdRm,
	},
	{
		Name:            "ssh",
		Usage:           "Log into or run a command on a machine with SSH.",
		Description:     "Arguments are [machine-name] [command]",
		Action:          cmdSsh,
		SkipFlagParsing: true,
	},
	{
		Name:        "scp",
		Usage:       "Copy files between machines",
		Description: "Arguments are [machine:][path] [machine:][path].",
		Action:      cmdScp,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "Copy files recursively (required to copy directories)",
			},
		},
	},
	{
		Name:        "start",
		Usage:       "Start a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdStart,
	},
	{
		Name:        "status",
		Usage:       "Get the status of a machine",
		Description: "Argument is a machine name.",
		Action:      cmdStatus,
	},
	{
		Name:        "stop",
		Usage:       "Stop a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdStop,
	},
	{
		Name:        "upgrade",
		Usage:       "Upgrade a machine to the latest version of Docker",
		Description: "Argument(s) are one or more machine names.",
		Action:      cmdUpgrade,
	},
	{
		Name:        "url",
		Usage:       "Get the URL of a machine",
		Description: "Argument is a machine name.",
		Action:      cmdUrl,
	},
}

func printIP(h *host.Host) func() error {
	return func() error {
		ip, err := h.Driver.GetIP()
		if err != nil {
			return fmt.Errorf("Error getting IP address: %s", err)
		}
		fmt.Println(ip)
		return nil
	}
}

// machineCommand maps the command name to the corresponding machine command.
// We run commands concurrently and communicate back an error if there was one.
func machineCommand(actionName string, host *host.Host, errorChan chan<- error) {
	// TODO: These actions should have their own type.
	commands := map[string](func() error){
		"configureAuth": host.ConfigureAuth,
		"start":         host.Start,
		"stop":          host.Stop,
		"restart":       host.Restart,
		"kill":          host.Kill,
		"upgrade":       host.Upgrade,
		"ip":            printIP(host),
	}

	log.Debugf("command=%s machine=%s", actionName, host.Name)

	if err := commands[actionName](); err != nil {
		errorChan <- err
		return
	}

	errorChan <- nil
}

// runActionForeachMachine will run the command across multiple machines
func runActionForeachMachine(actionName string, machines []*host.Host) {
	var (
		numConcurrentActions = 0
		serialMachines       = []*host.Host{}
		errorChan            = make(chan error)
	)

	for _, machine := range machines {
		// Virtualbox is temperamental about doing things concurrently,
		// so we schedule the actions in a "queue" to be executed serially
		// after the concurrent actions are scheduled.
		switch machine.DriverName {
		case "virtualbox":
			machine := machine
			serialMachines = append(serialMachines, machine)
		default:
			numConcurrentActions++
			go machineCommand(actionName, machine, errorChan)
		}
	}

	// While the concurrent actions are running,
	// do the serial actions.  As the name implies,
	// these run one at a time.
	for _, machine := range serialMachines {
		serialChan := make(chan error)
		go machineCommand(actionName, machine, serialChan)
		if err := <-serialChan; err != nil {
			log.Errorln(err)
		}
		close(serialChan)
	}

	// TODO: We should probably only do 5-10 of these
	// at a time, since otherwise cloud providers might
	// rate limit us.
	for i := 0; i < numConcurrentActions; i++ {
		if err := <-errorChan; err != nil {
			log.Errorln(err)
		}
	}

	close(errorChan)
}

func runActionWithContext(actionName string, c *cli.Context) error {
	store := getStore(c)

	hosts, err := getHostsFromContext(c)
	if err != nil {
		return err
	}

	if len(hosts) == 0 {
		log.Fatal(ErrNoMachineSpecified)
	}

	runActionForeachMachine(actionName, hosts)

	for _, h := range hosts {
		if err := store.Save(h); err != nil {
			return fmt.Errorf("Error saving host to store: %s", err)
		}
	}

	return nil
}

// Returns the cert paths.
// codegangsta/cli will not set the cert paths if the storage-path is set to
// something different so we cannot use the paths in the global options. le
// sigh.
func getCertPathInfoFromContext(c *cli.Context) cert.CertPathInfo {
	caCertPath := c.GlobalString("tls-ca-cert")
	caKeyPath := c.GlobalString("tls-ca-key")
	clientCertPath := c.GlobalString("tls-client-cert")
	clientKeyPath := c.GlobalString("tls-client-key")

	if caCertPath == "" {
		caCertPath = filepath.Join(mcndirs.GetMachineCertDir(), "ca.pem")
	}

	if caKeyPath == "" {
		caKeyPath = filepath.Join(mcndirs.GetMachineCertDir(), "ca-key.pem")
	}

	if clientCertPath == "" {
		clientCertPath = filepath.Join(mcndirs.GetMachineCertDir(), "cert.pem")
	}

	if clientKeyPath == "" {
		clientKeyPath = filepath.Join(mcndirs.GetMachineCertDir(), "key.pem")
	}

	return cert.CertPathInfo{
		CaCertPath:       caCertPath,
		CaPrivateKeyPath: caKeyPath,
		ClientCertPath:   clientCertPath,
		ClientKeyPath:    clientKeyPath,
	}
}

func detectShell() (string, error) {
	// check for windows env and not bash (i.e. msysgit, etc)
	// the SHELL env var is not set for processes in msysgit; we check
	// for TERM instead
	if runtime.GOOS == "windows" && os.Getenv("TERM") != "cygwin" {
		log.Printf("On Windows, please specify either 'cmd' or 'powershell' with the --shell flag.\n\n")
		return "", ErrUnknownShell
	}

	// attempt to get the SHELL env var
	shell := filepath.Base(os.Getenv("SHELL"))

	log.Debugf("shell: %s", shell)
	if shell == "" {
		return "", ErrUnknownShell
	}

	return shell, nil
}
