package main

import (
	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(vmwarefusion.NewDriver("", ""))
}
