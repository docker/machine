package main

import (
	"github.com/docker/machine/drivers/azure"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(azure.NewDriver("", ""))
}
