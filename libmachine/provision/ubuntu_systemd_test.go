package provision

import (
	"testing"
)

func TestUbuntuSystemdCompatibleWithHost(t *testing.T) {
	info := &OsRelease{
		ID:        "ubuntu",
		VersionID: "15.04",
	}
	p := NewUbuntuSystemdProvisioner(nil)
	p.SetOsReleaseInfo(info)

	compatible := p.CompatibleWithHost()

	if !compatible {
		t.Fatalf("expected to be compatible with ubuntu 15.04")
	}

	info.VersionID = "14.04"

	compatible = p.CompatibleWithHost()

	if compatible {
		t.Fatalf("expected to NOT be compatible with ubuntu 14.04")
	}

}
