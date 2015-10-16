package main

import (
	"github.com/docker/machine/drivers/rackspace"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(rackspace.NewDriver("", ""))
}
