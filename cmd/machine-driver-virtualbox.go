package main

import (
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(virtualbox.NewDriver("", ""))
}
