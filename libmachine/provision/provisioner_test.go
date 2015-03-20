package provision

import (
	"testing"
)

func TestRegisty(t *testing.T) {
	provisioners = make(map[string]*RegisteredProvisioner)

	Register("fake", &RegisteredProvisioner{
		New: func(sshCommand SSHCommandFunc) Provisioner {
			return &FakeProvisioner{}
		},
	})

	if len(provisioners) == 0 {
		t.Fatal("Not enough provisioners")
	}

	if len(provisioners) > 1 {
		t.Fatal("Too many provisioners")
	}

	fake := provisioners["fake"].New(nil)

	if fake == nil {
		t.Fatal("Unable to create fake provisioner")
	}
}

type FakeProvisioner struct {
	CompWithHost  bool
	OsReleaseInfo *OsRelease
}

func (provisioner *FakeProvisioner) Service(name string, action ServiceState) error {
	return nil
}

func (provisioner *FakeProvisioner) Package(name string, action PackageState) error {
	return nil
}

func (provisioner *FakeProvisioner) Hostname() (string, error) {
	return "", nil
}

func (provisioner *FakeProvisioner) SetHostname(hostname string) error {
	return nil
}

func (provisioner *FakeProvisioner) CompatibleWithHost() bool {
	return provisioner.CompWithHost
}

func (provisioner *FakeProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}
