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

package commands

import (
	//"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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
	// b, err := client.GetUserTokens()
	response, err := client.GetUserTokens()
	if err != nil {
		log.Errorf("Failed user tokens response %v", err)
		return
	}
	// response, err := api.GetTokensFromJSON(b)
	// if err != nil {
	// 	log.Errorf("Failed user tokens %v", err)
	// 	return
	// }
	log.Infof("User tokens:")
	for _, token := range response.Tokens {
		log.Infof("----------------------------------------------")
		token.Display()
	}
}

func doGetUserToken(c *cli.Context) {
	log.Infof("Get user token : %s", c.String("tokenid"))
	client := getClient(c)
	// b, err := client.GetUserToken(c.String("tokenid"))
	response, err := client.GetUserToken(c.String("tokenid"))
	if err != nil {
		log.Errorf("Failed user token response %v", err)
		return
	}
	// response, err := api.GetTokenFromJSON(b)
	// if err != nil {
	// 	log.Errorf("Failed user token  %v", err)
	// 	return
	// }
	log.Infof("Token: ")
	response.Token.Display()
}

func doCreateToken(c *cli.Context) {
	log.Infof("Create token %s %s %s",
		c.String("email"),
		c.String("password"),
		c.Bool("expires"))
	client := getClient(c)
	// b, err := client.CreateToken(
	// 	c.String("email"),
	// 	c.String("password"),
	// 	c.Bool("expires"))
	response, err := client.CreateToken(
		c.String("email"),
		c.String("password"),
		c.Bool("expires"))
	if err != nil {
		log.Errorf("Creating token: %v", err)
	}
	// response, err := api.GetTokenFromJSON(b)
	// if err != nil {
	// 	log.Errorf("Failed response %v", err)
	// 	return
	// }
	log.Infof("Token created: ")
	response.Token.Display()
}

func doDeleteToken(c *cli.Context) {
	log.Infof("Remove token %s", c.String("tokenid"))
	client := getClient(c)
	// b, err := client.DeleteToken(c.String("tokenid"))
	err := client.DeleteToken(c.String("tokenid"))
	if err != nil {
		log.Errorf("Retrieving token: %v", err)
	}
	log.Infof("Token deleted")
}

func doUpdateToken(c *cli.Context) {
	log.Infof("Update token expiration time %s", c.String("tokenid"))
	client := getClient(c)
	// b, err := client.UpdateToken(c.String("tokenid"))
	response, err := client.UpdateToken(c.String("tokenid"))
	if err != nil {
		log.Errorf("Retrieving token: %v", err)
	}
	// response, err := api.GetTokenFromJSON(b)
	// if err != nil {
	// 	log.Errorf("Failed response %v", err)
	// 	return
	// }
	log.Infof("Token updated: ")
	response.Token.Display()
}
