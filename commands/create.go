package commands

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/docker/machine/commands/mcndirs"
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
	"github.com/docker/machine/drivers/errdriver"
)

var (
	errNoMachineName = errors.New("Error: No machine name specified")
)

func cmdCreate(c CommandLine, store persist.Store) error {
	if len(c.Args()) > 1 {
		return fmt.Errorf("Invalid command line. Found extra arguments %v", c.Args()[1:])
	}

	name := c.Args().First()
	if name == "" {
		c.ShowHelp()

		driverName := c.String("driver")
		h, err := store.NewHost(name, driverName)
		if err != nil {
			return fmt.Errorf("Error getting new host: %q", err) // TODO
		}

		if _, ok := h.Driver.(*errdriver.Driver); ok {
			return errdriver.ErrDriverNotLoadable{driverName}
		}

		createFlags := h.Driver.GetCreateFlags()
		if len(createFlags) > 0 {
			// TODO: This is just a prototype
			//	sort.Sort(ByFlagName(cmd.Flags))
			//--virtualbox-disk-size "20000"									Size of disk for host in MB [$VIRTUALBOX_DISK_SIZE]
			//--virtualbox-hostonly-cidr "192.168.99.1/24"								Specify the Host Only CIDR [$VIRTUALBOX_HOSTONLY_CIDR]
			//--virtualbox-import-boot2docker-vm 									The name of a Boot2Docker VM to import
			//--virtualbox-memory "1024"										Size of memory for host in MB [$VIRTUALBOX_MEMORY_SIZE]

			fmt.Println("\nDriver specific options:\n")
			for _, f := range createFlags {
				fmt.Printf("   --%s \"%v\"\t%s [%s]\n", f.String(), f.Default(), f.Description(), f.EnvVarName())
			}

			fmt.Println()
		}

		return errNoMachineName
	}

	validName := host.ValidateHostName(name)
	if !validName {
		return fmt.Errorf("Error creating machine: %s", mcnerror.ErrInvalidHostname)
	}

	if err := validateSwarmDiscovery(c.String("swarm-discovery")); err != nil {
		return fmt.Errorf("Error parsing swarm discovery: %s", err)
	}

	exists, err := store.Exists(name)
	if err != nil {
		return fmt.Errorf("Error checking if host exists: %s", err)
	}
	if exists {
		return mcnerror.ErrHostAlreadyExists{
			Name: name,
		}
	}

	driverName := c.String("driver")
	h, err := store.NewHost(name, driverName)
	if err != nil {
		return fmt.Errorf("Error getting new host: %s", err)
	}

	certInfo := getCertPathInfoFromContext(c)
	h.HostOptions = &host.HostOptions{
		AuthOptions: &auth.AuthOptions{
			CertDir:          mcndirs.GetMachineCertDir(),
			CaCertPath:       certInfo.CaCertPath,
			CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
			ClientCertPath:   certInfo.ClientCertPath,
			ClientKeyPath:    certInfo.ClientKeyPath,
			ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), name, "server.pem"),
			ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), name, "server-key.pem"),
			StorePath:        filepath.Join(mcndirs.GetMachineDir(), name),
		},
		EngineOptions: &engine.EngineOptions{
			ArbitraryFlags:   c.StringSlice("engine-opt"),
			Env:              c.StringSlice("engine-env"),
			InsecureRegistry: c.StringSlice("engine-insecure-registry"),
			Labels:           c.StringSlice("engine-label"),
			RegistryMirror:   c.StringSlice("engine-registry-mirror"),
			StorageDriver:    c.String("engine-storage-driver"),
			TlsVerify:        true,
			InstallURL:       c.String("engine-install-url"),
		},
		SwarmOptions: &swarm.SwarmOptions{
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

	// driverOpts is the actual data we send over the wire to set the
	// driver parameters (an interface fulfilling drivers.DriverOptions,
	// concrete type rpcdriver.RpcFlags).
	mcnFlags := h.Driver.GetCreateFlags()
	driverOpts := getDriverOpts(c, mcnFlags)

	if err := h.Driver.SetConfigFromFlags(driverOpts); err != nil {
		return err
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
		if ok {
			driverOpts.Values[name] = getter.Get()
		} else {
			// TODO: This is pretty hacky.  StringSlice is the only
			// type so far we have to worry about which is not a
			// Getter, though.
			driverOpts.Values[name] = c.StringSlice(name)
		}
	}

	return driverOpts
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
