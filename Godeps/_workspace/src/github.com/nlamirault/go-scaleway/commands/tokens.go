// Copyright (C) 2015 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	//"fmt"

	//log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/nlamirault/go-scaleway/log"
)

var commandGetToken = cli.Command{
	Name:        "get",
	Usage:       "List an individual token",
	Description: ``,
	Action:      doGetUserToken,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "tokenid",
			Usage: "Token unique identifier",
			Value: "",
		},
	},
}

var commandCreateToken = cli.Command{
	Name:        "create",
	Usage:       "Create a token",
	Description: ``,
	Action:      doCreateToken,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "email",
			Usage: "The user email",
			Value: "",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "The user password",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "expires",
			Usage: "Set if you want a Token wich doesn’t expire",
		},
	},
}

var commandListTokens = cli.Command{
	Name:        "list",
	Usage:       "List all tokens associate with your account",
	Description: ``,
	Action:      doListUserTokens,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

var commandDeleteToken = cli.Command{
	Name:        "delete",
	Usage:       "Delete a token",
	Description: ``,
	Action:      doDeleteToken,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "tokenid",
			Usage: "Token unique identifier",
			Value: "",
		},
	},
}

var commandUpdateToken = cli.Command{
	Name:        "update",
	Usage:       "Update a token",
	Description: ``,
	Action:      doUpdateToken,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "tokenid",
			Usage: "Token unique identifier",
			Value: "",
		},
	},
}

func doListUserTokens(c *cli.Context) {
	log.Infof("List user tokens")
	client := getClient(c)
	response, err := client.GetUserTokens()
	if err != nil {
		log.Errorf("Failed user tokens response %v", err)
		return
	}
	log.Infof("User tokens:")
	for _, token := range response.Tokens {
		log.Infof("----------------------------------------------")
		token.Display()
	}
}

func doGetUserToken(c *cli.Context) {
	log.Infof("Get user token : %s", c.String("tokenid"))
	client := getClient(c)
	response, err := client.GetUserToken(c.String("tokenid"))
	if err != nil {
		log.Errorf("Failed user token response %v", err)
		return
	}
	response.Token.Display()
}

func doCreateToken(c *cli.Context) {
	log.Infof("Create token %s %s %s",
		c.String("email"), c.String("password"), c.Bool("expires"))
	client := getClient(c)
	response, err := client.CreateToken(
		c.String("email"), c.String("password"), c.Bool("expires"))
	if err != nil {
		log.Errorf("Creating token: %v", err)
		return
	}
	log.Infof("Token created: ")
	response.Token.Display()
}

func doDeleteToken(c *cli.Context) {
	log.Infof("Remove token %s", c.String("tokenid"))
	client := getClient(c)
	err := client.DeleteToken(c.String("tokenid"))
	if err != nil {
		log.Errorf("Retrieving token: %v", err)
		return
	}
	log.Infof("Token deleted")
}

func doUpdateToken(c *cli.Context) {
	log.Infof("Update token expiration time %s", c.String("tokenid"))
	client := getClient(c)
	response, err := client.UpdateToken(c.String("tokenid"))
	if err != nil {
		log.Errorf("Retrieving token: %v", err)
		return
	}
	log.Infof("Token updated: ")
	response.Token.Display()
}
