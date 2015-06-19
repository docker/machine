// Copyright (C) 2015  Nicolas Lamirault <nicolas.lamirault@gmail.com>

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

var commandListServers = cli.Command{
	Name:        "list",
	Usage:       "List all servers associate with your account",
	Description: ``,
	Action:      doListServers,
	Flags: []cli.Flag{
		verboseFlag,
	},
}
var commandListServerActions = cli.Command{
	Name:        "actions",
	Usage:       "List actions to be applied on a server",
	Description: ``,
	Action:      doListActions,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

var commandGetServer = cli.Command{
	Name:        "get",
	Usage:       "Retrieve a server",
	Description: ``,
	Action:      doGetServer,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "serverid",
			Usage: "Server unique identifier",
			Value: "",
		},
	},
}

var commandDeleteServer = cli.Command{
	Name:        "delete",
	Usage:       "Delete a server",
	Description: ``,
	Action:      doDeleteServer,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "serverid",
			Usage: "Server unique identifier",
			Value: "",
		},
	},
}

var commandActionServer = cli.Command{
	Name:        "action",
	Usage:       "Execute an action on a server",
	Description: "Execute an action on a server",
	Action:      doActionServer,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "serverid",
			Usage: "Server unique identifier",
			Value: "",
		},
		cli.StringFlag{
			Name:  "action",
			Usage: "the action to perform",
			Value: "",
		},
	},
}

func doListServers(c *cli.Context) {
	log.Infof("List servers")
	client := getClient(c)
	response, err := client.GetServers()
	if err != nil {
		log.Errorf("Retrieving servers %v", err)
		return
	}
	log.Infof("Servers: ")
	for _, server := range response.Servers {
		log.Infof("----------------------------------------------")
		server.Display()
	}
}

func doGetServer(c *cli.Context) {
	log.Infof("Getting server %s", c.String("serverid"))
	client := getClient(c)
	response, err := client.GetServer(c.String("serverid"))
	if err != nil {
		log.Errorf("Retrieving server: %v", err)
		return
	}
	log.Infof("Server: ")
	response.Server.Display()
}

func doDeleteServer(c *cli.Context) {
	log.Infof("Remove server %s", c.String("serverid"))
	client := getClient(c)
	err := client.DeleteServer(c.String("serverid"))
	if err != nil {
		log.Errorf("Retrieving server: %v", err)
		return
	}
	log.Infof("Server deleted")
}

func doListActions(c *cli.Context) {
	log.Infof("List actions to be applied on a server")
	client := getClient(c)
	response, err := client.ListServerActions(c.String("serverid"))
	if err != nil {
		log.Errorf("List server actions: %v", err)
		return
	}
	log.Infof("Actions: ")
	for _, action := range response.Actions {
		log.Infof("----------------------------------------------")
		log.Infof("Name: %s", action)
	}
}

func doActionServer(c *cli.Context) {
	log.Infof("Perform action %s on server %s",
		c.String("action"), c.String("serverid"))
	client := getClient(c)
	response, err := client.PerformServerAction(
		c.String("serverid"), c.String("action"))
	if err != nil {
		log.Errorf("Failed: %v", err)
		return
	}
	log.Infof("Task: ")
	log.Infof("Id          : %s", response.Task.ID)
	log.Infof("Status      : %s", response.Task.Status)
	log.Infof("Description : %s", response.Task.Description)

}
