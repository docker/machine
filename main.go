package main

import (
	"os"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/utils"
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
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CA_CERT",
			Name:   "tls-ca-cert",
			Usage:  "CA to verify remotes against",
			Value:  filepath.Join(utils.GetMachineDir(), "ca.pem"),
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CA_KEY",
			Name:   "tls-ca-key",
			Usage:  "Private key to generate certificates",
			Value:  filepath.Join(utils.GetMachineDir(), "key.pem"),
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CLIENT_CERT",
			Name:   "tls-client-cert",
			Usage:  "Client cert to use for TLS",
			Value:  filepath.Join(utils.GetMachineClientCertDir(), "cert.pem"),
		},
		cli.StringFlag{
			EnvVar: "MACHINE_TLS_CLIENT_KEY",
			Name:   "tls-client-key",
			Usage:  "Private key used in client TLS auth",
			Value:  filepath.Join(utils.GetMachineClientCertDir(), "key.pem"),
		},
	}

	app.Run(os.Args)
}
