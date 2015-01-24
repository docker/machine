package main

import (
	"os"
	"path"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/utils"
)

func before(c *cli.Context) error {

	caCertPath := c.GlobalString("tls-ca-cert")
	caKeyPath := c.GlobalString("tls-ca-key")
	clientCertPath := c.GlobalString("tls-client-cert")
	clientKeyPath := c.GlobalString("tls-client-key")
	org := "docker"
	bits := 2048

	if _, err := os.Stat(utils.GetMachineDir()); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(utils.GetMachineDir(), 0700); err != nil {
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
		log.Debugf("Creating client certificate: %s", clientCertPath)

		if _, err := os.Stat(utils.GetMachineClientCertDir()); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(utils.GetMachineClientCertDir(), 0700); err != nil {
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

		// copy ca.pem to client cert dir for docker client
		if err := utils.CopyFile(caCertPath, filepath.Join(utils.GetMachineClientCertDir(), "ca.pem")); err != nil {
			log.Fatalf("Error copying ca.pem to client cert dir: %s", err)
		}
	}

	return nil
}

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
	app.Before = before
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
