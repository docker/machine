package main

import (
	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(digitalocean.NewDriver("", ""))
}
