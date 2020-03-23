package provision

import (
	"testing"

	"github.com/rancher/machine/drivers/fakedriver"
	"github.com/rancher/machine/libmachine/auth"
	"github.com/rancher/machine/libmachine/engine"
	"github.com/rancher/machine/libmachine/provision/provisiontest"
	"github.com/rancher/machine/libmachine/swarm"
)

func TestRedHatDefaultStorageDriver(t *testing.T) {
	p := NewRedHatProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	if p.EngineOptions.StorageDriver != "overlay2" {
		t.Fatal("Default storage driver should be overlay2")
	}
}
