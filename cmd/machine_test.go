package machine

import (
	"testing"

	"github.com/docker/machine/commands/mcndirs"
)

func TestStorePathSetCorrectly(t *testing.T) {
	mcndirs.BaseDir = ""

	App().Run([]string{"docker-machine", "--storage-path", "/tmp/foo"})

	if mcndirs.BaseDir != "/tmp/foo" {
		t.Fatal("Expected BaseDir to be /tmp/foo but was ", mcndirs.BaseDir)
	}
}
