package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/docker/machine/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

var (
	ErrUnknownShell       = errors.New("Error: Unknown shell")
	ErrNoMachineSpecified = errors.New("Error: Expected to get one or more machine names as arguments")
	ErrExpectedOneMachine = errors.New("Error: Expected one machine name as an argument")
)

// CommandLine contains all the information passed to the commands on the command line.
type CommandLine interface {
	ShowHelp()

	Application() *cli.App

	Args() cli.Args

	Bool(name string) bool

	String(name string) string

	StringSlice(name string) []string

	GlobalString(name string) string

	FlagNames() (names []string)

	Generic(name string) interface{}
}

type contextCommandLine struct {
	*cli.Context
}

func (c *contextCommandLine) ShowHelp() {
	cli.ShowCommandHelp(c.Context, c.Command.Name)
}

func (c *contextCommandLine) Application() *cli.App {
	return c.App
}

func fatalOnError(command func(commandLine CommandLine, store persist.Store) error) func(context *cli.Context) {
	return func(context *cli.Context) {
		commandLine := &contextCommandLine{context}
		store := getStore(commandLine)

		if err := command(commandLine, store); err != nil {
			log.Fatal(err)
		}
	}
}

func getStore(c CommandLine) persist.Store {
	certInfo := getCertPathInfoFromContext(c)
	return &persist.Filestore{
		Path:             c.GlobalString("storage-path"),
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	}
}

func confirmInput(msg string) (bool, error) {
	fmt.Printf("%s (y/n): ", msg)

	var resp string
	_, err := fmt.Scanln(&resp)
	if err != nil {
		return false, err
	}

	confirmed := strings.Index(strings.ToLower(resp), "y") == 0
	return confirmed, nil
}

func saveHost(store persist.Store, h *host.Host) error {
	if err := store.Save(h); err != nil {
		return fmt.Errorf("Error attempting to save host to store: %s", err)
	}

	return nil
}

var Commands = []cli.Command{
	{
		Name:   "active",
		Usage:  "Print which machine is active",
		Action: fatalOnError(cmdActive),
	},
	{
		Name:        "config",
		Usage:       "Print the connection config for machine",
		Description: "Argument is a machine name.",
		Action:      fatalOnError(cmdConfig),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "swarm",
				Usage: "Display the Swarm config instead of the Docker daemon",
			},
		},
	},
	{
		Name:   "create",
		Usage:  fmt.Sprintf("Create a machine.\n\nRun '%s create --driver name' to include the create flags for that driver in the help text.", os.Args[0]),
		Action: fatalOnError(cmdCreate),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "driver, d",
				Usage: fmt.Sprintf(
					"Driver to create machine with.",
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
		},
	},
	{
		Name:        "env",
		Usage:       "Display the commands to set up the environment for the Docker client",
		Description: "Argument is a machine name.",
		Action:      fatalOnError(cmdEnv),
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
		Action:      fatalOnError(cmdInspect),
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
		Action:      fatalOnError(cmdIP),
	},
	{
		Name:        "kill",
		Usage:       "Kill a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      fatalOnError(cmdKill),
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
		Action: fatalOnError(cmdLs),
	},
	{
		Name:        "regenerate-certs",
		Usage:       "Regenerate TLS Certificates for a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      fatalOnError(cmdRegenerateCerts),
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
		Action:      fatalOnError(cmdRestart),
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
		Action:      fatalOnError(cmdRm),
	},
	{
		Name:            "ssh",
		Usage:           "Log into or run a command on a machine with SSH.",
		Description:     "Arguments are [machine-name] [command]",
		Action:          fatalOnError(cmdSSH),
		SkipFlagParsing: true,
	},
	{
		Name:        "scp",
		Usage:       "Copy files between machines",
		Description: "Arguments are [machine:][path] [machine:][path].",
		Action:      fatalOnError(cmdScp),
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
		Action:      fatalOnError(cmdStart),
	},
	{
		Name:        "status",
		Usage:       "Get the status of a machine",
		Description: "Argument is a machine name.",
		Action:      fatalOnError(cmdStatus),
	},
	{
		Name:        "stop",
		Usage:       "Stop a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      fatalOnError(cmdStop),
	},
	{
		Name:        "upgrade",
		Usage:       "Upgrade a machine to the latest version of Docker",
		Description: "Argument(s) are one or more machine names.",
		Action:      fatalOnError(cmdUpgrade),
	},
	{
		Name:        "url",
		Usage:       "Get the URL of a machine",
		Description: "Argument is a machine name.",
		Action:      fatalOnError(cmdURL),
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

	errorChan <- commands[actionName]()
}

// runActionForeachMachine will run the command across multiple machines
func runActionForeachMachine(actionName string, machines []*host.Host) []error {
	var (
		numConcurrentActions = 0
		serialMachines       = []*host.Host{}
		errorChan            = make(chan error)
		errs                 = []error{}
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
			errs = append(errs, err)
		}
		close(serialChan)
	}

	// TODO: We should probably only do 5-10 of these
	// at a time, since otherwise cloud providers might
	// rate limit us.
	for i := 0; i < numConcurrentActions; i++ {
		if err := <-errorChan; err != nil {
			errs = append(errs, err)
		}
	}

	close(errorChan)

	return errs
}

func consolidateErrs(errs []error) error {
	finalErr := ""
	for _, err := range errs {
		finalErr = fmt.Sprintf("%s\n%s", finalErr, err)
	}

	return errors.New(strings.TrimSpace(finalErr))
}

func runActionWithContext(actionName string, c CommandLine, store persist.Store) error {
	hosts := []*host.Host{}

	if len(c.Args()) == 0 {
		return ErrNoMachineSpecified
	}

	for _, hostName := range c.Args() {
		h, err := store.Load(hostName)
		if err != nil {
			return fmt.Errorf("Could not load host %q: %s", hostName, err)
		}
		hosts = append(hosts, h)
	}

	if errs := runActionForeachMachine(actionName, hosts); len(errs) > 0 {
		return consolidateErrs(errs)
	}

	for _, h := range hosts {
		if err := saveHost(store, h); err != nil {
			return fmt.Errorf("Error saving host to store: %s", err)
		}
	}

	return nil
}

// Returns the cert paths.
// codegangsta/cli will not set the cert paths if the storage-path is set to
// something different so we cannot use the paths in the global options. le
// sigh.
func getCertPathInfoFromContext(c CommandLine) cert.CertPathInfo {
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
	// attempt to get the SHELL env var
	shell := filepath.Base(os.Getenv("SHELL"))

	log.Debugf("shell: %s", shell)
	if shell == "" {
		// check for windows env and not bash (i.e. msysgit, etc)
		if runtime.GOOS == "windows" {
			log.Printf("On Windows, please specify either 'cmd' or 'powershell' with the --shell flag.\n\n")
		}

		return "", ErrUnknownShell
	}

	return shell, nil
}
