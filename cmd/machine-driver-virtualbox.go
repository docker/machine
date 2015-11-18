package main

import (
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/docker/machine/libmachine/log"
)

func main() {
	plugin.RegisterDriver(virtualbox.NewDriver("", ""))
	log.Debug("debug from virtualbox")
	log.Debugf("debugf from virtualbox (%s)", "foo")
	log.Print("print from virtualbox")
	log.Printf("printf from virtualbox (%s)", "foo")
	log.Error("error from virtualbox")
	log.Errorf("errorf from virtualbox (%s)", "foo")
	log.Warn("warnfrom virtualbox")
	log.Warnf("warnf from virtualbox (%s)", "foo")
	log.Info("info from virtualbox")
	log.Infof("infof from virtualbox (%s)", "foo")
	log.Fatal("fatal from virtualbox")
	log.Fatalf("fatalf from virtualbox (%s)", "foo")
}
