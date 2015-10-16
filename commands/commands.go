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
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/persist"
)

var (
	ErrUnknownShell       = errors.New("Error: Unknown shell")
	ErrNoMachineSpecified = errors.New("Error: Expected to get one or more machine names as arguments.")
	ErrExpectedOneMachine = errors.New("Error: Expected one machine name as an argument.")

	RpcClientDriversCh = make(chan *rpcdriver.RpcClientDriver)
	RpcDriversClosedCh = make(chan bool)
)

func newPluginDriver(driverName string, rawContent []byte) (*rpcdriver.RpcClientDriver, error) {
	d, err := rpcdriver.NewRpcClientDriver(rawContent, driverName)
	if err != nil {
		return nil, err
	}

	RpcClientDriversCh <- d

	return d, nil
}

func DeferClosePluginServers() {
	rpcClientDrivers := []*rpcdriver.RpcClientDriver{}

	for d := range RpcClientDriversCh {
		rpcClientDrivers = append(rpcClientDrivers, d)
	}

	doneCh := make(chan bool)

	for _, d := range rpcClientDrivers {
		d := d
		go func() {
			if err := d.Close(); err != nil {
				log.Debugf("Error closing connection to plugin server: %s", err)
			}

			doneCh <- true
		}()
	}

	for range rpcClientDrivers {
		<-doneCh
	}

	RpcDriversClosedCh <- true
}

func fatal(args ...interface{}) {
	close(RpcClientDriversCh)
	<-RpcDriversClosedCh
	log.Fatal(args...)
}

func fatalf(fmtString string, args ...interface{}) {
	close(RpcClientDriversCh)
	<-RpcDriversClosedCh
	log.Fatalf(fmtString, args...)
}

func confirmInput(msg string) bool {
	fmt.Printf("%s (y/n): ", msg)
	var resp string
	_, err := fmt.Scanln(&resp)

	if err != nil {
		fatal(err)
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

func listHosts(store persist.Store) ([]*host.Host, error) {
	cliHosts := []*host.Host{}

	hosts, err := store.List()
	if err != nil {
		return nil, fmt.Errorf("Error attempting to list hosts from store: %s", err)
	}

	for _, h := range hosts {
		d, err := newPluginDriver(h.DriverName, h.RawDriver)
		if err != nil {
			return nil, fmt.Errorf("Error attempting to invoke binary for plugin: %s", err)
		}

		h.Driver = d

		cliHosts = append(cliHosts, h)
	}

	return cliHosts, nil
}

func loadHost(store persist.Store, hostName string) (*host.Host, error) {
	h, err := store.Load(hostName)
	if err != nil {
		return nil, fmt.Errorf("Loading host from store failed: %s", err)
	}

	d, err := newPluginDriver(h.DriverName, h.RawDriver)
	if err != nil {
		return nil, fmt.Errorf("Error attempting to invoke binary for plugin: %s", err)
	}

	h.Driver = d

	return h, nil
}

func saveHost(store persist.Store, h *host.Host) error {
	if err := store.Save(h); err != nil {
		return fmt.Errorf("Error attempting to save host to store: %s", err)
	}

	return nil
}

func getFirstArgHost(c *cli.Context) *host.Host {
	store := getStore(c)
	hostName := c.Args().First()

	exists, err := store.Exists(hostName)
	if err != nil {
		fatalf("Error checking if host %q exists: %s", hostName, err)
	}

	if !exists {
		fatalf("Host %q does not exist", hostName)
	}

	h, err := loadHost(store, hostName)
	if err != nil {
		// I guess I feel OK with bailing here since if we can't get
		// the host reliably we're definitely not going to be able to
		// do anything else interesting, but also this premature exit
		// feels wrong to me.  Let's revisit it later.
		fatalf("Error trying to get host %q: %s", hostName, err)
	}
	return h
}

func getHostsFromContext(c *cli.Context) ([]*host.Host, error) {
	store := getStore(c)
	hosts := []*host.Host{}

	for _, hostName := range c.Args() {
		h, err := loadHost(store, hostName)
		if err != nil {
			return nil, fmt.Errorf("Could not load host %q: %s", hostName, err)
		}
		hosts = append(hosts, h)
	}

	return hosts, nil
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
		Flags:           sharedCreateFlags,
		Name:            "create",
		Usage:           "Create a machine.\n\nSpecify a driver with --driver to include the create flags for that driver in the help text.",
		Action:          cmdCreateOuter,
		SkipFlagParsing: true,
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

func runActionWithContext(actionName string, c *cli.Context) error {
	store := getStore(c)

	hosts, err := getHostsFromContext(c)
	if err != nil {
		return err
	}

	if len(hosts) == 0 {
		return ErrNoMachineSpecified
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
