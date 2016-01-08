package provision

import (
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
)

func TestDebianDefaultStorageDriver(t *testing.T) {
	p := NewDebianProvisioner(&fakedriver.Driver{}).(*DebianProvisioner)
	p.SSHCommander = provisiontest.FakeSSHCommander{FilesystemType: "ext4"}
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	if p.EngineOptions.StorageDriver != "aufs" {
		t.Fatal("Default storage driver should be aufs")
	}
}
