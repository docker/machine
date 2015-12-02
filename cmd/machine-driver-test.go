package main

import (
	"github.com/docker/machine/drivers/test"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(test.NewDriver("", ""))
}
