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
	//"log"
	//"os"

	//log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/nlamirault/go-scaleway/log"
)

var commandListIPs = cli.Command{
	Name:        "list",
	Usage:       "List availables IPs",
	Description: ``,
	Action:      doListIPs,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

var commandCreateIP = cli.Command{
	Name:        "create",
	Usage:       "Create an IP",
	Description: ``,
	Action:      doCreateIP,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

var commandGetIP = cli.Command{
	Name:        "get",
	Usage:       "Retrieve an IP",
	Description: ``,
	Action:      doGetIP,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "ipid",
			Usage: "IP unique identifier",
			Value: "",
		},
	},
}

var commandDeleteIP = cli.Command{
	Name:        "delete",
	Usage:       "Delete an IP",
	Description: ``,
	Action:      doDeleteIP,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "ipid",
			Usage: "IP unique identifier",
			Value: "",
		},
	},
}

func doCreateIP(c *cli.Context) {
	log.Infof("Create IP")
	client := getClient(c)
	response, err := client.CreateIP()
	if err != nil {
		log.Errorf("Create IP: %v", err)
		return
	}
	log.Infof("IP: ")
	response.IPAddress.Display()
}

func doGetIP(c *cli.Context) {
	log.Infof("Getting IP %s", c.String("ipid"))
	client := getClient(c)
	response, err := client.GetIP(c.String("ipid"))
	if err != nil {
		log.Errorf("Retrieving IP: %s", err.Error())
		return
	}
	log.Infof("IP: ")
	response.IPAddress.Display()
}

func doDeleteIP(c *cli.Context) {
	log.Infof("Remove IP %s", c.String("IPid"))
	client := getClient(c)
	err := client.DeleteIP(c.String("IPid"))
	if err != nil {
		log.Errorf("Retrieving IP: %v", err)
		return
	}
	log.Infof("IP deleted")
}

func doListIPs(c *cli.Context) {
	log.Infof("List IPs")
	client := getClient(c)
	response, err := client.GetIPs()
	if err != nil {
		log.Errorf("Retrieving IPs %v", err)
		return
	}
	log.Infof("IPs: ")
	for _, ip := range response.IPAddresses {
		log.Infof("----------------------------------------------")
		ip.Display()
	}
}
