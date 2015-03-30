package provision

import (
	"reflect"
	"testing"
)

func TestParseOsRelease(t *testing.T) {
	// These example osr files stolen shamelessly from
	// https://github.com/docker/docker/blob/master/pkg/parsers/operatingsystem/operatingsystem_test.go
	// cheers @tiborvass
	var (
		ubuntuTrusty = []byte(`NAME="Ubuntu"
VERSION="14.04, Trusty Tahr"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 14.04 LTS"
VERSION_ID="14.04"
HOME_URL="http://www.ubuntu.com/"
SUPPORT_URL="http://help.ubuntu.com/"
BUG_REPORT_URL="http://bugs.launchpad.net/ubuntu/"`)
		gentoo = []byte(`NAME=Gentoo
ID=gentoo
PRETTY_NAME="Gentoo/Linux"
ANSI_COLOR="1;32"
HOME_URL="http://www.gentoo.org/"
SUPPORT_URL="http://www.gentoo.org/main/en/support.xml"
BUG_REPORT_URL="https://bugs.gentoo.org/"
`)
		noPrettyName = []byte(`NAME="Ubuntu"
VERSION="14.04, Trusty Tahr"
ID=ubuntu
ID_LIKE=debian
VERSION_ID="14.04"
HOME_URL="http://www.ubuntu.com/"
SUPPORT_URL="http://help.ubuntu.com/"
BUG_REPORT_URL="http://bugs.launchpad.net/ubuntu/"`)
	)

	osr, err := NewOsRelease(ubuntuTrusty)
	if err != nil {
		t.Fatal("Unexpected error parsing os release: %s", err)
	}

	expectedOsr := OsRelease{
		AnsiColor:    "",
		Name:         "Ubuntu",
		Version:      "14.04, Trusty Tahr",
		Id:           "ubuntu",
		IdLike:       "debian",
		PrettyName:   "Ubuntu 14.04 LTS",
		VersionId:    "14.04",
		HomeUrl:      "http://www.ubuntu.com/",
		SupportUrl:   "http://help.ubuntu.com/",
		BugReportUrl: "http://bugs.launchpad.net/ubuntu/",
	}

	if !reflect.DeepEqual(*osr, expectedOsr) {
		t.Fatal("Error with ubuntu osr parsing: structs do not match")
	}

	osr, err = NewOsRelease(gentoo)
	if err != nil {
		t.Fatal("Unexpected error parsing os release: %s", err)
	}

	expectedOsr = OsRelease{
		AnsiColor:    "1;32",
		Name:         "Gentoo",
		Version:      "",
		Id:           "gentoo",
		IdLike:       "",
		PrettyName:   "Gentoo/Linux",
		VersionId:    "",
		HomeUrl:      "http://www.gentoo.org/",
		SupportUrl:   "http://www.gentoo.org/main/en/support.xml",
		BugReportUrl: "https://bugs.gentoo.org/",
	}

	if !reflect.DeepEqual(*osr, expectedOsr) {
		t.Fatal("Error with gentoo osr parsing: structs do not match")
	}

	osr, err = NewOsRelease(noPrettyName)
	if err != nil {
		t.Fatal("Unexpected error parsing os release: %s", err)
	}

	expectedOsr = OsRelease{
		AnsiColor:    "",
		Name:         "Ubuntu",
		Version:      "14.04, Trusty Tahr",
		Id:           "ubuntu",
		IdLike:       "debian",
		PrettyName:   "",
		VersionId:    "14.04",
		HomeUrl:      "http://www.ubuntu.com/",
		SupportUrl:   "http://help.ubuntu.com/",
		BugReportUrl: "http://bugs.launchpad.net/ubuntu/",
	}

	if !reflect.DeepEqual(*osr, expectedOsr) {
		t.Fatal("Error with noPrettyName osr parsing: structs do not match")
	}
}

func TestParseLine(t *testing.T) {
	var (
		withQuotes    = "ID=\"ubuntu\""
		withoutQuotes = "ID=gentoo"
		wtf           = "LOTS=OF=EQUALS"
	)

	key, val, err := parseLine(withQuotes)
	if key != "ID" {
		t.Fatalf("Expected ID, got %s", key)
	}
	if val != "ubuntu" {
		t.Fatalf("Expected ubuntu, got %s", val)
	}
	if err != nil {
		t.Fatalf("Got error on parseLine with quotes: %s", err)
	}
	key, val, err = parseLine(withoutQuotes)
	if key != "ID" {
		t.Fatalf("Expected ID, got %s", key)
	}
	if val != "gentoo" {
		t.Fatalf("Expected gentoo, got %s", val)
	}
	if err != nil {
		t.Fatalf("Got error on parseLine without quotes: %s", err)
	}
	key, val, err = parseLine(wtf)
	if err == nil {
		t.Fatal("Expected to get an error on parseLine, got nil")
	}
}
