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

var commandGetUser = cli.Command{
	Name:        "get",
	Usage:       "List informations about your user account",
	Description: ``,
	Action:      doListUserInformations,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

var commandListOrganizations = cli.Command{
	Name:        "list",
	Usage:       "List all Organizations associate with your account",
	Description: ``,
	Action:      doListUserOrganizations,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

func doListUserInformations(c *cli.Context) {
	log.Debugf("List user informations")
	client := getClient(c)
	response, err := client.GetUserInformations()
	if err != nil {
		log.Errorf("Failed user response %v", err)
		return
	}
	log.Infof("User: ")
	response.User.Display()
}

func doListUserOrganizations(c *cli.Context) {
	log.Debugf("List user organizations")
	client := getClient(c)
	response, err := client.GetUserOrganizations()
	if err != nil {
		log.Errorf("Failed user organizations response %v", err)
		return
	}
	log.Infof("Organizations:")
	for _, org := range response.Organizations {
		log.Infof("----------------------------------------------")
		org.Display()
	}
}
