package main

import (
	"github.com/docker/machine/drivers/google"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(google.NewDriver("", ""))
}
