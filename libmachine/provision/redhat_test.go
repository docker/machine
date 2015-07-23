package provision

import (
	"regexp"
	"testing"
)

func TestRedHatGenerateYumRepoList(t *testing.T) {
	info := &OsRelease{
		Id: "rhel",
	}
	p := NewRedHatProvisioner(nil)
	p.SetOsReleaseInfo(info)

	buf, err := generateYumRepoList(p)
	if err != nil {
		t.Fatal(err)
	}

	m, err := regexp.MatchString(".*centos/7.*", buf.String())
	if err != nil {
		t.Fatal(err)
	}

	if !m {
		t.Fatalf("expected match for centos/7")
	}
}
