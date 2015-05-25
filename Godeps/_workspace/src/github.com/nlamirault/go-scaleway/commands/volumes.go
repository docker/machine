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

var commandListVolumes = cli.Command{
	Name:        "list",
	Usage:       "List availables volumes",
	Description: ``,
	Action:      doListVolumes,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

var commandGetVolume = cli.Command{
	Name:        "get",
	Usage:       "Retrieve a volume",
	Description: ``,
	Action:      doGetVolume,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "volumeid",
			Usage: "Volume unique identifier",
			Value: "",
		},
	},
}

var commandDeleteVolume = cli.Command{
	Name:        "delete",
	Usage:       "Delete a volume",
	Description: ``,
	Action:      doDeleteVolume,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "volumeid",
			Usage: "Volume unique identifier",
			Value: "",
		},
	},
}

var commandCreateVolume = cli.Command{
	Name:        "create",
	Usage:       "Create a volume",
	Description: ``,
	Action:      doCreateVolume,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "name",
			Usage: "The volume name",
			Value: "",
		},
		cli.StringFlag{
			Name:  "type",
			Usage: "The volume type (l_hdd|l_ssd)",
			Value: "",
		},
		cli.IntFlag{
			Name:  "size",
			Usage: "The volume size",
			Value: 0,
		},
	},
}

func doListVolumes(c *cli.Context) {
	log.Infof("List volumes")
	client := getClient(c)
	response, err := client.GetVolumes()
	if err != nil {
		log.Errorf("Retrieving volumes %v", err)
		return
	}
	log.Infof("Volumes: ")
	for _, volume := range response.Volumes {
		log.Infof("----------------------------------------------")
		volume.Display()
	}
}

func doGetVolume(c *cli.Context) {
	log.Infof("Getting volume %s", c.String("volumeid"))
	client := getClient(c)
	response, err := client.GetVolume(c.String("volumeid"))
	if err != nil {
		log.Errorf("Retrieving volume: %v", err)
		return
	}
	log.Infof("Volume: ")
	response.Volume.Display()
}

func doDeleteVolume(c *cli.Context) {
	log.Infof("Remove volume %s", c.String("volumeid"))
	client := getClient(c)
	err := client.DeleteVolume(c.String("volumeid"))
	if err != nil {
		log.Errorf("Retrieving volume: %v", err)
		return
	}
	log.Infof("Volume deleted")
}

func doCreateVolume(c *cli.Context) {
	log.Infof("Create volume %s %s %s %s",
		c.String("name"), c.String("type"), c.Int("size"))
	client := getClient(c)
	response, err := client.CreateVolume(
		c.String("name"), c.String("type"), c.Int("size"))
	if err != nil {
		log.Errorf("Creating volume: %v", err)
		return
	}
	log.Infof("Volume: ")
	response.Volume.Display()
}
