package main

import (
	"github.com/docker/machine/drivers/vmwarevsphere"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(vmwarevsphere.NewDriver("", ""))
}
