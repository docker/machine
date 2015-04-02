// Copyright (C) 2015  Nicolas Lamirault <nicolas.lamirault@gmail.com>

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
	//"log"
	//"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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
			Name:  "organizationid",
			Usage: "Organization unique identifier",
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
	//b, err := client.GetVolumes()
	response, err := client.GetVolumes()
	if err != nil {
		log.Errorf("Retrieving volumes %v", err)
		return
	}
	// response, err := api.GetVolumesFromJSON(b)
	// if err != nil {
	// 	log.Errorf("Reading volumes %v", err)
	// 	return
	// }
	log.Infof("Volumes: ")
	for _, volume := range response.Volumes {
		log.Infof("----------------------------------------------")
		volume.Display()
	}
}

func doGetVolume(c *cli.Context) {
	log.Infof("Getting volume %s", c.String("volumeid"))
	client := getClient(c)
	// b, err := client.GetVolume(c.String("volumeid"))
	response, err := client.GetVolume(c.String("volumeid"))
	if err != nil {
		log.Errorf("Retrieving volume: %v", err)
	}
	// response, err := api.GetVolumeFromJSON(b)
	// if err != nil {
	// 	log.Errorf("Failed response %v", err)
	// 	return
	// }
	log.Infof("Volume: ")
	response.Volume.Display()
}

func doDeleteVolume(c *cli.Context) {
	log.Infof("Remove volume %s", c.String("volumeid"))
	client := getClient(c)
	err := client.DeleteVolume(c.String("volumeid"))
	if err != nil {
		log.Errorf("Retrieving volume: %v", err)
	}
	log.Infof("Volume deleted")
}

func doCreateVolume(c *cli.Context) {
	log.Infof("Create volume %s %s %s %s",
		c.String("name"),
		c.String("organizationid"),
		c.String("type"),
		c.Int("size"))
	client := getClient(c)
	// b, err := client.CreateVolume(
	// 	c.String("name"),
	// 	c.String("organizationid"),
	// 	c.String("type"),
	// 	c.Int("size"))
	response, err := client.CreateVolume(
		c.String("name"),
		c.String("organizationid"),
		c.String("type"),
		c.Int("size"))
	if err != nil {
		log.Errorf("Creating volume: %v", err)
	}
	// response, err := api.GetVolumeFromJSON(b)
	// if err != nil {
	// 	log.Errorf("Failed response %v", err)
	// 	return
	// }
	log.Infof("Volume: ")
	response.Volume.Display()
}
