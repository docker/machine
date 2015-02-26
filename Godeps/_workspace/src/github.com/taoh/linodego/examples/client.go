// +build example
package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/taoh/linodego"
	"os"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel) // set debug level
	}

	apiKey := "[API Key]"
	client := linodego.NewClient(apiKey, nil)

	// Create a linode
	v, err := client.Linode.Create(2, 1, 1)
	if err != nil {
		log.Fatal(err)
	}
	linodeId := v.LinodeId.LinodeId
	log.Infof("Created linode: %d", linodeId)

	// Get IP Address
	v2, err := client.Ip.List(linodeId, -1)
	if err != nil {
		log.Fatal(err)
	}
	fullIpAddress := v2.FullIPAddresses[0]
	log.Infof("IP Address: Id: %d, Address: %s", fullIpAddress.IPAddressId, fullIpAddress.IPAddress)

	// Update label, https://github.com/taoh/linodego/issues/1
	args := make(map[string]interface{})
	args["Label"] = 12345
	_, err = client.Linode.Update(linodeId, args)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Updated Linode: Label %s", args)

	// List linode
	v3, err := client.Linode.List(linodeId)
	if err != nil {
		log.Fatal(err)
	}
	linode := v3.Linodes[0]
	log.Infof("List Linode: %d, Label: %s, Status: %d", linode.LinodeId, linode.Label, linode.Status)

	// Shutdown the linode
	v4, err := client.Linode.Shutdown(linodeId)
	if err != nil {
		log.Fatal(err)
	}
	job := v4.Job
	log.Infof("Shutdown linode: %s", job)

	// Delete the linode
	_, err = client.Linode.Delete(linodeId, false)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Deleted linode: %s", linodeId)
}
