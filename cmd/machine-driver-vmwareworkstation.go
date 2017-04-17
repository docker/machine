package main

import (
	"github.com/docker/machine/drivers/vmwareworkstation"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(vmwareworkstation.NewDriver("", ""))
}
