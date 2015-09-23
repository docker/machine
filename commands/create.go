package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/driverfactory"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/swarm"
)

var (
	ErrDriverNotRecognized = errors.New("Driver not recognized.")
)

func cmdCreate(c *cli.Context) {
	var (
		driver drivers.Driver
	)

	driverName := c.String("driver")
	name := c.Args().First()
	certInfo := getCertPathInfoFromContext(c)

	storePath := c.GlobalString("storage-path")
	store := &persist.Filestore{
		Path:             storePath,
		CaCertPath:       certInfo.CaCertPath,
		CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
	}

	// TODO: Not really a fan of "none" as the default driver...
	if driverName != "none" {
		var err error

		c.App.Commands, err = trimDriverFlags(driverName, c.App.Commands)
		if err != nil {
			log.Fatal(err)
		}
	}

	if name == "" {
		cli.ShowCommandHelp(c, "create")
		log.Fatal("You must specify a machine name")
	}

	validName := host.ValidateHostName(name)
	if !validName {
		log.Fatal("Error creating machine: ", mcnerror.ErrInvalidHostname)
	}

	if err := validateSwarmDiscovery(c.String("swarm-discovery")); err != nil {
		log.Fatalf("Error parsing swarm discovery: %s", err)
	}

	hostOptions := &host.HostOptions{
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

	driver, err := driverfactory.NewDriver(driverName, name, storePath)
	if err != nil {
		log.Fatalf("Error trying to get driver: %s", err)
	}

	h, err := store.NewHost(driver)
	if err != nil {
		log.Fatalf("Error getting new host: %s", err)
	}

	h.HostOptions = hostOptions

	exists, err := store.Exists(h.Name)
	if err != nil {
		log.Fatalf("Error checking if host exists: %s", err)
	}
	if exists {
		log.Fatal(mcnerror.ErrHostAlreadyExists{
			Name: h.Name,
		})
	}

	// TODO: This should be moved out of the driver and done in the
	// commands module.
	if err := h.Driver.SetConfigFromFlags(c); err != nil {
		log.Fatalf("Error setting machine configuration from flags provided: %s", err)
	}

	if err := libmachine.Create(store, h); err != nil {
		log.Fatalf("Error creating machine: %s", err)
	}

	info := fmt.Sprintf("%s env %s", os.Args[0], name)
	log.Infof("To see how to connect Docker to this machine, run: %s", info)
}

// If the user has specified a driver, they should not see the flags for all
// of the drivers in `docker-machine create`.  This method replaces the 100+
// create flags with only the ones applicable to the driver specified
func trimDriverFlags(driver string, cmds []cli.Command) ([]cli.Command, error) {
	filteredCmds := cmds
	driverFlags, err := drivers.GetCreateFlagsForDriver(driver)
	if err != nil {
		return nil, err
	}

	for i, cmd := range cmds {
		if cmd.HasName("create") {
			filteredCmds[i].Flags = append(driverFlags, sharedCreateFlags...)
		}
	}

	return filteredCmds, nil
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
