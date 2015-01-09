package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	for _, f := range os.Args {
		if f == "-D" || f == "--debug" || f == "-debug" {
			os.Setenv("DEBUG", "1")
			initLogging(log.DebugLevel)
		}
	}

	app := cli.NewApp()
	app.Name = "machine"
	app.Commands = Commands
	app.CommandNotFound = cmdNotFound
	app.Usage = "Create and manage machines running Docker."
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug mode",
		},
		cli.StringFlag{
			EnvVar: "MACHINE_STORAGE_PATH",
			Name:   "storage-path",
			Usage:  "Configures storage path",
		},
	}

	app.Run(os.Args)
}
