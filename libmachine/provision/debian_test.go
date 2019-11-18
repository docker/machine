package provision

import (
	"testing"

	"github.com/rancher/machine/drivers/fakedriver"
	"github.com/rancher/machine/libmachine/auth"
	"github.com/rancher/machine/libmachine/engine"
	"github.com/rancher/machine/libmachine/provision/provisiontest"
	"github.com/rancher/machine/libmachine/swarm"
)

func TestDebianDefaultStorageDriver(t *testing.T) {
	p := NewDebianProvisioner(&fakedriver.Driver{}).(*DebianProvisioner)
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	if p.EngineOptions.StorageDriver != "aufs" {
		t.Fatal("Default storage driver should be aufs")
	}
}
