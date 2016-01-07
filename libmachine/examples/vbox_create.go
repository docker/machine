package main

// Sample Virtualbox create independent of Machine CLI.
import (
	"encoding/json"
	"fmt"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func main() {
	log.SetDebug(true)

	client := libmachine.NewClient("/tmp/automatic", "/tmp/automatic/certs")
	defer client.Close()

	hostName := "myfunhost"

	// Set some options on the provider...
	driver := virtualbox.NewDriver(hostName, "/tmp/automatic")
	driver.CPU = 2
	driver.Memory = 2048

	data, err := json.Marshal(driver)
	if err != nil {
		log.Error(err)
		return
	}

	h, err := client.NewHost("virtualbox", data)
	if err != nil {
		log.Error(err)
		return
	}

	h.HostOptions.EngineOptions.StorageDriver = "overlay"

	if err := client.Create(h); err != nil {
		log.Error(err)
		return
	}

	out, err := h.RunSSHCommand("df -h")
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Printf("Results of your disk space query:\n%s\n", out)

	fmt.Println("Powering down machine now...")
	if err := h.Stop(); err != nil {
		log.Error(err)
		return
	}
}
