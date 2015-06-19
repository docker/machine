package commands

import (
	"fmt"
	"path/filepath"

	"github.com/docker/machine/log"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/utils"
)

func init() {
	buildCmdCreate()
}

func buildCmdCreate() {
	cmd := cli.Command{
		Name:  "create",
		Usage: "Create a machine",
	}
	subCmds := []cli.Command{}
	names := drivers.GetDriverNames()
	var err error
	for _, name := range names {
		subCmd := cli.Command{
			Name:   name,
			Usage:  "Create a machine using the " + name + " driver",
			Action: cmdCreate,
		}
		// if there is an error here there is something really wrong
		subCmd.Flags, err = drivers.GetCreateFlagsForDriver(name)
		if err != nil {
			panic(err)
		}
		subCmd.Flags = append(subCmd.Flags, sharedCreateFlags...)
		subCmds = append(subCmds, subCmd)
	}
	cmd.Subcommands = subCmds
	Commands = append(Commands, cmd)
}

func cmdCreate(c *cli.Context) {
	driver := c.Command.Name
	name := c.Args().First()

	if name == "" {
		cli.ShowCommandHelp(c, driver)
		log.Fatal("You must specify a machine name")
	}

	certInfo := getCertPathInfo(c)

	if err := setupCertificates(
		certInfo.CaCertPath,
		certInfo.CaKeyPath,
		certInfo.ClientCertPath,
		certInfo.ClientKeyPath); err != nil {
		log.Fatalf("Error generating certificates: %s", err)
	}

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

	hostOptions := &libmachine.HostOptions{
		AuthOptions: &auth.AuthOptions{
			CaCertPath:     certInfo.CaCertPath,
			PrivateKeyPath: certInfo.CaKeyPath,
			ClientCertPath: certInfo.ClientCertPath,
			ClientKeyPath:  certInfo.ClientKeyPath,
			ServerCertPath: filepath.Join(utils.GetMachineDir(), name, "server.pem"),
			ServerKeyPath:  filepath.Join(utils.GetMachineDir(), name, "server-key.pem"),
		},
		EngineOptions: &engine.EngineOptions{
			ArbitraryFlags:   c.StringSlice("engine-opt"),
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

	_, err = provider.Create(name, driver, hostOptions, c)
	if err != nil {
		log.Errorf("Error creating machine: %s", err)
		log.Fatal("You will want to check the provider to make sure the machine and associated resources were properly removed.")
	}

	info := fmt.Sprintf("%s env %s", c.App.Name, name)
	log.Infof("To see how to connect Docker to this machine, run: %s", info)
}
