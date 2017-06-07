package provision

import (
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
)

func TestRaspbianDefaultStorageDriver(t *testing.T) {
	p := NewRaspbianProvisioner(&fakedriver.Driver{}).(*RaspbianProvisioner)
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	if p.EngineOptions.StorageDriver != "overlay" {
		t.Fatal("Default storage driver should be overlay")
	}
}
