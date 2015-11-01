package main

import (
	"github.com/docker/machine/drivers/vmwarevcloudair"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(vmwarevcloudair.NewDriver("", ""))
}
