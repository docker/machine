package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"errors"

	"github.com/docker/machine/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/errdriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/swarm"
)

var (
	errNoMachineName = errors.New("Error: No machine name specified")
)

var (
	sharedCreateFlags = []cli.Flag{
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
	}
)

func cmdCreateInner(c CommandLine) error {
	if len(c.Args()) > 1 {
		return fmt.Errorf("Invalid command line. Found extra arguments %v", c.Args()[1:])
	}

	name := c.Args().First()
	driverName := c.String("driver")
	certInfo := getCertPathInfoFromContext(c)

	storePath := c.GlobalString("storage-path")

	store := &persist.Filestore{
		Path:             storePath,
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	}

	if name == "" {
		c.ShowHelp()
		return errNoMachineName
	}

	validName := host.ValidateHostName(name)
	if !validName {
		return fmt.Errorf("Error creating machine: %s", mcnerror.ErrInvalidHostname)
	}

	if err := validateSwarmDiscovery(c.String("swarm-discovery")); err != nil {
		return fmt.Errorf("Error parsing swarm discovery: %s", err)
	}

	// TODO: Fix hacky JSON solution
	bareDriverData, err := json.Marshal(&drivers.BaseDriver{
		MachineName: name,
		StorePath:   c.GlobalString("storage-path"),
	})
	if err != nil {
		return fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	driver, err := newPluginDriver(driverName, bareDriverData)
	if err != nil {
		return fmt.Errorf("Error loading driver %q: %s", driverName, err)
	}

	h, err := store.NewHost(driver)
	if err != nil {
		return fmt.Errorf("Error getting new host: %s", err)
	}

	h.HostOptions = &host.Options{
		AuthOptions: &auth.Options{
			CertDir:          mcndirs.GetMachineCertDir(),
			CaCertPath:       certInfo.CaCertPath,
			CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
			ClientCertPath:   certInfo.ClientCertPath,
			ClientKeyPath:    certInfo.ClientKeyPath,
			ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), name, "server.pem"),
			ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), name, "server-key.pem"),
			StorePath:        filepath.Join(mcndirs.GetMachineDir(), name),
		},
		EngineOptions: &engine.Options{
			ArbitraryFlags:   c.StringSlice("engine-opt"),
			Env:              c.StringSlice("engine-env"),
			InsecureRegistry: c.StringSlice("engine-insecure-registry"),
			Labels:           c.StringSlice("engine-label"),
			RegistryMirror:   c.StringSlice("engine-registry-mirror"),
			StorageDriver:    c.String("engine-storage-driver"),
			TLSVerify:        true,
			InstallURL:       c.String("engine-install-url"),
		},
		SwarmOptions: &swarm.Options{
			IsSwarm:        c.Bool("swarm"),
			Image:          c.String("swarm-image"),
			Master:         c.Bool("swarm-master"),
			Discovery:      c.String("swarm-discovery"),
			Address:        c.String("swarm-addr"),
			Host:           c.String("swarm-host"),
			Strategy:       c.String("swarm-strategy"),
			ArbitraryFlags: c.StringSlice("swarm-opt"),
		},
	}

	exists, err := store.Exists(h.Name)
	if err != nil {
		return fmt.Errorf("Error checking if host exists: %s", err)
	}
	if exists {
		return mcnerror.ErrHostAlreadyExists{
			Name: h.Name,
		}
	}

	// driverOpts is the actual data we send over the wire to set the
	// driver parameters (an interface fulfilling drivers.DriverOptions,
	// concrete type rpcdriver.RpcFlags).
	mcnFlags := driver.GetCreateFlags()
	driverOpts := getDriverOpts(c, mcnFlags)

	if err := h.Driver.SetConfigFromFlags(driverOpts); err != nil {
		return fmt.Errorf("Error setting machine configuration from flags provided: %s", err)
	}

	if err := libmachine.Create(store, h); err != nil {
		return fmt.Errorf("Error creating machine: %s", err)
	}

	if err := saveHost(store, h); err != nil {
		return fmt.Errorf("Error attempting to save store: %s", err)
	}

	log.Infof("To see how to connect Docker to this machine, run: %s", fmt.Sprintf("%s env %s", os.Args[0], name))

	return nil
}

// The following function is needed because the CLI acrobatics that we're doing
// (with having an "outer" and "inner" function each with their own custom
// settings and flag parsing needs) are not well supported by codegangsta/cli.
//
// Instead of trying to make a convoluted series of flag parsing and relying on
// codegangsta/cli internals work well, we simply read the flags we're
// interested in from the outer function into module-level variables, and then
// use them from the "inner" function.
//
// I'm not very pleased about this, but it seems to be the only decent
// compromise without drastically modifying codegangsta/cli internals or our
// own CLI.
func flagHackLookup(flagName string) string {
	// e.g. "-d" for "--driver"
	flagPrefix := flagName[1:3]

	// TODO: Should we support -flag-name (single hyphen) syntax as well?
	for i, arg := range os.Args {
		if strings.Contains(arg, flagPrefix) {
			// format '--driver foo' or '-d foo'
			if arg == flagPrefix || arg == flagName {
				if i+1 < len(os.Args) {
					return os.Args[i+1]
				}
			}

			// format '--driver=foo' or '-d=foo'
			if strings.HasPrefix(arg, flagPrefix+"=") || strings.HasPrefix(arg, flagName+"=") {
				return strings.Split(arg, "=")[1]
			}
		}
	}

	return ""
}

func cmdCreateOuter(c CommandLine) error {
	const (
		flagLookupMachineName = "flag-lookup"
	)
	driverName := flagHackLookup("--driver")

	// We didn't recognize the driver name.
	if driverName == "" {
		c.ShowHelp()
		return nil // ?
	}

	// TODO: Fix hacky JSON solution
	bareDriverData, err := json.Marshal(&drivers.BaseDriver{
		MachineName: flagLookupMachineName,
	})
	if err != nil {
		return fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	driver, err := newPluginDriver(driverName, bareDriverData)
	if err != nil {
		return fmt.Errorf("Error loading driver %q: %s", driverName, err)
	}

	if _, ok := driver.(*errdriver.Driver); ok {
		return errdriver.ErrDriverNotLoadable{driverName}
	}

	// TODO: So much flag manipulation and voodoo here, it seems to be
	// asking for trouble.
	//
	// mcnFlags is the data we get back over the wire (type mcnflag.Flag)
	// to indicate which parameters are available.
	mcnFlags := driver.GetCreateFlags()

	// This bit will actually make "create" display the correct flags based
	// on the requested driver.
	cliFlags, err := convertMcnFlagsToCliFlags(mcnFlags)
	if err != nil {
		return fmt.Errorf("Error trying to convert provided driver flags to cli flags: %s", err)
	}

	for i := range c.Application().Commands {
		cmd := &c.Application().Commands[i]
		if cmd.HasName("create") {
			cmd = addDriverFlagsToCommand(cliFlags, cmd)
		}
	}

	if serialDriver, ok := driver.(*drivers.SerialDriver); ok {
		driver = serialDriver.Driver
	}

	if rpcd, ok := driver.(*rpcdriver.RpcClientDriver); ok {
		if err := rpcd.Close(); err != nil {
			return err
		}
	}

	return c.Application().Run(os.Args)
}

func getDriverOpts(c CommandLine, mcnflags []mcnflag.Flag) drivers.DriverOptions {
	// TODO: This function is pretty damn YOLO and would benefit from some
	// sanity checking around types and assertions.
	//
	// But, we need it so that we can actually send the flags for creating
	// a machine over the wire (cli.Context is a no go since there is so
	// much stuff in it).
	driverOpts := rpcdriver.RpcFlags{
		Values: make(map[string]interface{}),
	}

	for _, f := range mcnflags {
		driverOpts.Values[f.String()] = f.Default()

		// Hardcoded logic for boolean... :(
		if f.Default() == nil {
			driverOpts.Values[f.String()] = false
		}
	}

	for _, name := range c.FlagNames() {
		getter, ok := c.Generic(name).(flag.Getter)
		if !ok {
			// TODO: This is pretty hacky.  StringSlice is the only
			// type so far we have to worry about which is not a
			// Getter, though.
			driverOpts.Values[name] = c.StringSlice(name)
			continue
		}
		driverOpts.Values[name] = getter.Get()
	}

	return driverOpts
}

func convertMcnFlagsToCliFlags(mcnFlags []mcnflag.Flag) ([]cli.Flag, error) {
	cliFlags := []cli.Flag{}
	for _, f := range mcnFlags {
		switch t := f.(type) {
		// TODO: It seems pretty wrong to just default "nil" to this,
		// but cli.BoolFlag doesn't have a "Value" field (false is
		// always the default)
		case *mcnflag.BoolFlag:
			f := f.(*mcnflag.BoolFlag)
			cliFlags = append(cliFlags, cli.BoolFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
			})
		case *mcnflag.IntFlag:
			f := f.(*mcnflag.IntFlag)
			cliFlags = append(cliFlags, cli.IntFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
				Value:  f.Value,
			})
		case *mcnflag.StringFlag:
			f := f.(*mcnflag.StringFlag)
			cliFlags = append(cliFlags, cli.StringFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
				Value:  f.Value,
			})
		case *mcnflag.StringSliceFlag:
			f := f.(*mcnflag.StringSliceFlag)
			cliFlags = append(cliFlags, cli.StringSliceFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,

				//TODO: Is this used with defaults? Can we convert the literal []string to cli.StringSlice properly?
				Value: &cli.StringSlice{},
			})
		default:
			log.Warn("Flag is ", f)
			return nil, fmt.Errorf("Flag is unrecognized flag type: %T", t)
		}
	}

	return cliFlags, nil
}

func addDriverFlagsToCommand(cliFlags []cli.Flag, cmd *cli.Command) *cli.Command {
	cmd.Flags = append(sharedCreateFlags, cliFlags...)
	cmd.SkipFlagParsing = false
	cmd.Action = fatalOnError(cmdCreateInner)
	sort.Sort(ByFlagName(cmd.Flags))

	return cmd
}

func validateSwarmDiscovery(discovery string) error {
	if discovery == "" {
		return nil
	}

	matched, err := regexp.MatchString(`[^:]*://.*`, discovery)
	if err != nil {
		return err
	}

	if matched {
		return nil
	}

	return fmt.Errorf("Swarm Discovery URL was in the wrong format: %s", discovery)
}
