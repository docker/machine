package commands

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/utils"
)

func cmdCreate(c *cli.Context) {
	var (
		err error
	)
	driver := c.String("driver")
	name := c.Args().First()

	// TODO: Not really a fan of "none" as the default driver...
	if driver != "none" {
		c.App.Commands, err = trimDriverFlags(driver, c.App.Commands)
		if err != nil {
			log.Fatal(err)
		}
	}

	if name == "" {
		cli.ShowCommandHelp(c, "create")
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

	mcn, err := newMcn(defaultStore)
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
		EngineOptions: &engine.EngineOptions{},
		SwarmOptions: &swarm.SwarmOptions{
			IsSwarm:   c.Bool("swarm"),
			Master:    c.Bool("swarm-master"),
			Discovery: c.String("swarm-discovery"),
			Address:   c.String("swarm-addr"),
			Host:      c.String("swarm-host"),
		},
	}

	host, err := mcn.Create(name, driver, hostOptions, c)
	if err != nil {
		log.Errorf("Error creating machine: %s", err)
		log.Warn("You will want to check the provider to make sure the machine and associated resources were properly removed.")
		log.Fatal("Error creating machine")
	}
	if err := mcn.SetActive(host); err != nil {
		log.Fatalf("error setting active host: %v", err)
	}

	info := ""
	userShell := filepath.Base(os.Getenv("SHELL"))

	switch userShell {
	case "fish":
		info = fmt.Sprintf("%s env %s | source", c.App.Name, name)
	default:
		info = fmt.Sprintf(`eval "$(%s env %s)"`, c.App.Name, name)
	}

	log.Infof("%q has been created and is now the active machine.", name)

	if info != "" {
		log.Infof("To point your Docker client at it, run this in your shell: %s", info)
	}
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
