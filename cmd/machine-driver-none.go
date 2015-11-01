package main

import (
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(none.NewDriver("", ""))
}
