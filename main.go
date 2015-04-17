package main

import (
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/docker/machine/commands"
	"github.com/docker/machine/utils"
	"github.com/docker/machine/version"
)

func main() {
	for _, f := range os.Args {
		if f == "-D" || f == "--debug" || f == "-debug" {
			os.Setenv("DEBUG", "1")
			initLogging(log.DebugLevel)
		}
	}

	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Docker Machine Contributors"
	app.Email = "https://github.com/docker/machine"
	app.Before = func(c *cli.Context) error {
		os.Setenv("MACHINE_STORAGE_PATH", c.GlobalString("storage-path"))
		return nil
	}
	app.Commands = commands.Commands
	app.CommandNotFound = cmdNotFound
	app.Usage = "Create and manage machines running Docker."
	app.Version = version.VERSION + " (" + version.GITCOMMIT + ")"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug mode",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_STORAGE_PATH",
			Name:   "storage-path",
			Value:  utils.GetBaseDir(),
			Usage:  "Configures storage path",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CA_CERT",
			Name:   "tls-ca-cert",
			Usage:  "CA to verify remotes against",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CA_KEY",
			Name:   "tls-ca-key",
			Usage:  "Private key to generate certificates",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CLIENT_CERT",
			Name:   "tls-client-cert",
			Usage:  "Client cert to use for TLS",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CLIENT_KEY",
			Name:   "tls-client-key",
			Usage:  "Private key used in client TLS auth",
			Value:  "",
		},
	}

	app.Run(os.Args)
}

func cmdNotFound(c *cli.Context, command string) {
	log.Fatalf(
		"%s: '%s' is not a %s command. See '%s --help'.",
		c.App.Name,
		command,
		c.App.Name,
		c.App.Name,
	)
}
