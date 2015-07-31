package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/skarademir/naturalsort"

	"github.com/docker/machine/drivers"
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
	_ "github.com/docker/machine/drivers/vmwareappcatalyst"
	_ "github.com/docker/machine/drivers/vmwarefusion"
	_ "github.com/docker/machine/drivers/vmwarevcloudair"
	_ "github.com/docker/machine/drivers/vmwarevsphere"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/log"
	"github.com/docker/machine/utils"
)

var (
	ErrUnknownShell       = errors.New("Error: Unknown shell")
	ErrNoMachineSpecified = errors.New("Error: Expected to get one or more machine names as arguments.")
	ErrExpectedOneMachine = errors.New("Error: Expected one machine name as an argument.")
)

type machineConfig struct {
	machineName    string
	machineDir     string
	machineUrl     string
	clientKeyPath  string
	serverCertPath string
	clientCertPath string
	caCertPath     string
	caKeyPath      string
	serverKeyPath  string
	AuthOptions    auth.AuthOptions
	SwarmOptions   swarm.SwarmOptions
}

func sortHostListItemsByName(items []libmachine.HostListItem) {
	m := make(map[string]libmachine.HostListItem, len(items))
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

func newProvider(store libmachine.Store) (*libmachine.Provider, error) {
	return libmachine.New(store)
}

func getMachineDir(rootPath string) string {
	return filepath.Join(rootPath, "machines")
}

func getDefaultStore(rootPath, caCertPath, privateKeyPath string) (libmachine.Store, error) {
	return libmachine.NewFilestore(
		rootPath,
		caCertPath,
		privateKeyPath,
	), nil
}

func setupCertificates(caCertPath, caKeyPath, clientCertPath, clientKeyPath string) error {
	org := utils.GetUsername()
	bits := 2048

	if _, err := os.Stat(utils.GetMachineCertDir()); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(utils.GetMachineCertDir(), 0700); err != nil {
				log.Fatalf("Error creating machine config dir: %s", err)
			}
		} else {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		log.Infof("Creating CA: %s", caCertPath)

		// check if the key path exists; if so, error
		if _, err := os.Stat(caKeyPath); err == nil {
			log.Fatalf("The CA key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := utils.GenerateCACertificate(caCertPath, caKeyPath, org, bits); err != nil {
			log.Infof("Error generating CA certificate: %s", err)
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Infof("Creating client certificate: %s", clientCertPath)

		if _, err := os.Stat(utils.GetMachineCertDir()); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(utils.GetMachineCertDir(), 0700); err != nil {
					log.Fatalf("Error creating machine client cert dir: %s", err)
				}
			} else {
				log.Fatal(err)
			}
		}

		// check if the key path exists; if so, error
		if _, err := os.Stat(clientKeyPath); err == nil {
			log.Fatalf("The client key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := utils.GenerateCert([]string{""}, clientCertPath, clientKeyPath, caCertPath, caKeyPath, org, bits); err != nil {
			log.Fatalf("Error generating client certificate: %s", err)
		}
	}

	return nil
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
		Name:        "ssh",
		Usage:       "Log into or run a command on a machine with SSH.",
		Description: "Arguments are [machine-name] [command]",
		Action:      cmdSsh,
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

// machineCommand maps the command name to the corresponding machine command.
// We run commands concurrently and communicate back an error if there was one.
func machineCommand(actionName string, host *libmachine.Host, errorChan chan<- error) {
	commands := map[string](func() error){
		"configureAuth": host.ConfigureAuth,
		"start":         host.Start,
		"stop":          host.Stop,
		"restart":       host.Restart,
		"kill":          host.Kill,
		"upgrade":       host.Upgrade,
		"ip":            host.PrintIP,
	}

	log.Debugf("command=%s machine=%s", actionName, host.Name)

	if err := commands[actionName](); err != nil {
		errorChan <- err
		return
	}

	errorChan <- nil
}

// runActionForeachMachine will run the command across multiple machines
func runActionForeachMachine(actionName string, machines []*libmachine.Host) {
	var (
		numConcurrentActions = 0
		serialMachines       = []*libmachine.Host{}
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
	machines, err := getHosts(c)
	if err != nil {
		return err
	}

	if len(machines) == 0 {
		log.Fatal(ErrNoMachineSpecified)
	}

	runActionForeachMachine(actionName, machines)

	return nil
}

func getHosts(c *cli.Context) ([]*libmachine.Host, error) {
	machines := []*libmachine.Host{}
	for _, n := range c.Args() {
		machine, err := loadMachine(n, c)
		if err != nil {
			return nil, err
		}

		machines = append(machines, machine)
	}

	return machines, nil
}

func loadMachine(name string, c *cli.Context) (*libmachine.Host, error) {
	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := newProvider(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	host, err := provider.Get(name)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func getHost(c *cli.Context) *libmachine.Host {
	name := c.Args().First()

	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		c.GlobalString("tls-ca-cert"),
		c.GlobalString("tls-ca-key"),
	)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := newProvider(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	host, err := provider.Get(name)
	if err != nil {
		log.Fatalf("unable to load host: %v", err)
	}
	return host
}

func getDefaultProvider(c *cli.Context) *libmachine.Provider {
	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := newProvider(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	return provider
}

func getMachineConfig(c *cli.Context) (*machineConfig, error) {
	name := c.Args().First()
	certInfo := getCertPathInfo(c)
	defaultStore, err := getDefaultStore(
		c.GlobalString("storage-path"),
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := newProvider(defaultStore)
	if err != nil {
		log.Fatal(err)
	}

	m, err := provider.Get(name)
	if err != nil {
		return nil, err
	}

	machineDir := filepath.Join(utils.GetMachineDir(), m.Name)
	caCert := filepath.Join(machineDir, "ca.pem")
	caKey := filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	clientCert := filepath.Join(machineDir, "cert.pem")
	clientKey := filepath.Join(machineDir, "key.pem")
	serverCert := filepath.Join(machineDir, "server.pem")
	serverKey := filepath.Join(machineDir, "server-key.pem")
	machineUrl, err := m.GetURL()
	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			machineUrl = ""
		} else {
			return nil, fmt.Errorf("Unexpected error getting machine url: %s", err)
		}
	}
	return &machineConfig{
		machineName:    name,
		machineDir:     machineDir,
		machineUrl:     machineUrl,
		clientKeyPath:  clientKey,
		clientCertPath: clientCert,
		serverCertPath: serverCert,
		caKeyPath:      caKey,
		caCertPath:     caCert,
		serverKeyPath:  serverKey,
		AuthOptions:    *m.HostOptions.AuthOptions,
		SwarmOptions:   *m.HostOptions.SwarmOptions,
	}, nil
}

// getCertPaths returns the cert paths
// codegangsta/cli will not set the cert paths if the storage-path
// is set to something different so we cannot use the paths
// in the global options. le sigh.
func getCertPathInfo(c *cli.Context) libmachine.CertPathInfo {
	// setup cert paths
	caCertPath := c.GlobalString("tls-ca-cert")
	caKeyPath := c.GlobalString("tls-ca-key")
	clientCertPath := c.GlobalString("tls-client-cert")
	clientKeyPath := c.GlobalString("tls-client-key")

	if caCertPath == "" {
		caCertPath = filepath.Join(utils.GetMachineCertDir(), "ca.pem")
	}

	if caKeyPath == "" {
		caKeyPath = filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	}

	if clientCertPath == "" {
		clientCertPath = filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	}

	if clientKeyPath == "" {
		clientKeyPath = filepath.Join(utils.GetMachineCertDir(), "key.pem")
	}

	return libmachine.CertPathInfo{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
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
