package main

import (
	"github.com/docker/machine/drivers/softlayer"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(softlayer.NewDriver("", ""))
}
