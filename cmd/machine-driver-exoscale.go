package main

import (
	"github.com/docker/machine/drivers/exoscale"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(exoscale.NewDriver("", ""))
}
