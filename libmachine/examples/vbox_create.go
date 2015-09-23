package main

// Sample Virtualbox create independent of Machine CLI.
import (
	"fmt"
	"os"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func main() {
	libmachine.SetDebug(true)

	log.SetOutWriter(os.Stdout)
	log.SetErrWriter(os.Stderr)

	// returns the familiar store at $HOME/.docker/machine
	store := libmachine.GetDefaultStore()

	// over-ride this for now (don't want to muck with my default store)
	store.Path = "/tmp/automatic"

	hostName := "myfunhost"

	// Set some options on the provider...
	driver := virtualbox.NewDriver(hostName, "/tmp/automatic")
	driver.CPU = 2
	driver.Memory = 2048

	h, err := store.NewHost(driver)
	if err != nil {
		log.Fatal(err)
	}

	h.HostOptions.EngineOptions.StorageDriver = "overlay"

	if err := libmachine.Create(store, h); err != nil {
		log.Fatal(err)
	}

	out, err := h.RunSSHCommand("df -h")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Results of your disk space query:\n%s\n", out)

	fmt.Println("Powering down machine now...")
	if err := h.Stop(); err != nil {
		log.Fatal(err)
	}
}
