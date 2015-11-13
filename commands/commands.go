package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/machine/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/drivers/rpc"
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

func fatalOnError(command func(commandLine CommandLine, store rpcdriver.Store) error) func(context *cli.Context) {
	return func(context *cli.Context) {
		cli := &contextCommandLine{context}
		store := getStore(cli)

		if err := command(cli, store); err != nil {
			log.Fatal(err)
		}
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

func getStore(cli CommandLine) rpcdriver.Store {
	certInfo := getCertPathInfoFromContext(cli)

	return rpcdriver.NewRPCStore(&persist.Filestore{
		Path:             cli.GlobalString("storage-path"),
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	})
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
		Flags:           sharedCreateFlags,
		Name:            "create",
		Usage:           fmt.Sprintf("Create a machine.\n\nRun '%s create --driver name' to include the create flags for that driver in the help text.", os.Args[0]),
		Action:          fatalOnError(cmdCreateOuter),
		SkipFlagParsing: true,
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

// runActionForeachMachine will run the command across multiple machines
func runActionForeachMachine(action func(h *host.Host) error, machines []*host.Host) []error {
	errorChan := make(chan error)
	for _, machine := range machines {
		theMachine := machine

		go func() {
			errorChan <- action(theMachine)
		}()
	}

	// TODO: We should probably only do 5-10 of these
	// at a time, since otherwise cloud providers might
	// rate limit us.
	errs := []error{}
	for range machines {
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

func runActionOnHosts(action func(h *host.Host) error, store rpcdriver.Store, hostNames []string) error {
	if len(hostNames) == 0 {
		return ErrNoMachineSpecified
	}

	hosts, err := getHostsByName(store, hostNames)
	if err != nil {
		return err
	}

	if errs := runActionForeachMachine(action, hosts); len(errs) > 0 {
		return consolidateErrs(errs)
	}

	for _, h := range hosts {
		if err := store.Save(h); err != nil {
			return fmt.Errorf("Error saving host to store: %s", err)
		}
	}

	return nil
}

func getHostsByName(store rpcdriver.Store, hostNames []string) ([]*host.Host, error) {
	hosts := []*host.Host{}

	for _, hostName := range hostNames {
		h, err := store.Load(hostName)
		if err != nil {
			return nil, fmt.Errorf("Could not load host %q: %s", hostName, err)
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
}

// Returns the cert paths.
// codegangsta/cli will not set the cert paths if the storage-path is set to
// something different so we cannot use the paths in the global options. le
// sigh.
func getCertPathInfoFromContext(cli CommandLine) cert.PathInfo {
	caCertPath := cli.GlobalString("tls-ca-cert")
	if caCertPath == "" {
		caCertPath = filepath.Join(mcndirs.GetMachineCertDir(), "ca.pem")
	}

	caKeyPath := cli.GlobalString("tls-ca-key")
	if caKeyPath == "" {
		caKeyPath = filepath.Join(mcndirs.GetMachineCertDir(), "ca-key.pem")
	}

	clientCertPath := cli.GlobalString("tls-client-cert")
	if clientCertPath == "" {
		clientCertPath = filepath.Join(mcndirs.GetMachineCertDir(), "cert.pem")
	}

	clientKeyPath := cli.GlobalString("tls-client-key")
	if clientKeyPath == "" {
		clientKeyPath = filepath.Join(mcndirs.GetMachineCertDir(), "key.pem")
	}

	return cert.PathInfo{
		CaCertPath:       caCertPath,
		CaPrivateKeyPath: caKeyPath,
		ClientCertPath:   clientCertPath,
		ClientKeyPath:    clientKeyPath,
	}
}
