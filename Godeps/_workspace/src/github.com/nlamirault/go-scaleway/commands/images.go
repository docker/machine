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

var commandListImages = cli.Command{
	Name:        "list",
	Usage:       "List availables images",
	Description: ``,
	Action:      doListImages,
	Flags: []cli.Flag{
		verboseFlag,
	},
}

var commandGetImage = cli.Command{
	Name:        "get",
	Usage:       "Retrieve a image",
	Description: ``,
	Action:      doGetImage,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "imageid",
			Usage: "Image unique identifier",
			Value: "",
		},
	},
}

var commandDeleteImage = cli.Command{
	Name:        "delete",
	Usage:       "Delete an image",
	Description: ``,
	Action:      doDeleteImage,
	Flags: []cli.Flag{
		verboseFlag,
		cli.StringFlag{
			Name:  "imageid",
			Usage: "Image unique identifier",
			Value: "",
		},
	},
}

func doGetImage(c *cli.Context) {
	log.Infof("Getting image %s", c.String("imageid"))
	client := getClient(c)
	response, err := client.GetImage(c.String("imageid"))
	if err != nil {
		log.Errorf("Retrieving image: %v", err)
		return
	}
	log.Infof("Image: ")
	response.Image.Display()
}

func doDeleteImage(c *cli.Context) {
	log.Infof("Remove image %s", c.String("imageid"))
	client := getClient(c)
	err := client.DeleteImage(c.String("imageid"))
	if err != nil {
		log.Errorf("Retrieving image: %v", err)
		return
	}
	log.Infof("Image deleted")
}

func doListImages(c *cli.Context) {
	log.Infof("List images")
	client := getClient(c)
	response, err := client.GetImages()
	if err != nil {
		log.Errorf("Retrieving images %v", err)
		return
	}
	log.Infof("Images: ")
	for _, image := range response.Images {
		log.Infof("----------------------------------------------")
		image.Display()
	}
}
