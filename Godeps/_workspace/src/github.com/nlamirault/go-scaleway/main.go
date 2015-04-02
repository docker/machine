// Copyright (C) 2015 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package go-scaleway provides a CLI for the Scaleway cloud.

// For a full guide visit http://github.com/nlamirault/go-scaleway

package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/nlamirault/go-scaleway/commands"
	"github.com/nlamirault/go-scaleway/version"
)

func makeApp() *cli.App {
	app := cli.NewApp()
	app.Name = "scaleway"
	app.Version = version.Version
	app.Usage = "A CLI for Scaleway"
	app.Author = "Nicolas Lamirault"
	app.Email = "nicolas.lamirault@gmail.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level, l",
			Value: "info",
			Usage: fmt.Sprintf("Log level (options: debug, info, warn, error, fatal, panic)"),
		},
		cli.StringFlag{
			Name:   "scaleway-userid",
			Usage:  "Scaleway UserID",
			Value:  "",
			EnvVar: "SCALEWAY_USERID",
		},
		cli.StringFlag{
			Name:   "scaleway-token",
			Usage:  "Scaleway Token",
			Value:  "",
			EnvVar: "SCALEWAY_TOKEN",
		},
		cli.StringFlag{
			Name:   "scaleway-organization",
			Usage:  "Organization identifier",
			Value:  "",
			EnvVar: "SCALEWAY_ORGANIZATION",
		},
	}
	app.Before = func(c *cli.Context) error {
		//log.SetFormatter(&logging.CustomFormatter{})
		log.SetOutput(os.Stderr)
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.SetLevel(level)
		return nil
	}

	app.Commands = commands.Commands
	//app.Flags = commands.Flags
	return app
}

func main() {
	app := makeApp()
	app.Run(os.Args)
}
