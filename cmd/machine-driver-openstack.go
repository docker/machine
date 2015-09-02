package main

import (
	"github.com/docker/machine/drivers/openstack"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(openstack.NewDriver("", ""))
}
