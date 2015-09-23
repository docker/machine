package mcndirs

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/mcnutils"
)

func TestGetBaseDir(t *testing.T) {
	// reset any override env var
	homeDir := mcnutils.GetHomeDir()
	baseDir := GetBaseDir()

	if strings.Index(baseDir, homeDir) != 0 {
		t.Fatalf("expected base dir with prefix %s; received %s", homeDir, baseDir)
	}
}

func TestGetCustomBaseDir(t *testing.T) {
	root := "/tmp"
	os.Setenv("MACHINE_STORAGE_PATH", root)
	baseDir := GetBaseDir()

	if strings.Index(baseDir, root) != 0 {
		t.Fatalf("expected base dir with prefix %s; received %s", root, baseDir)
	}
	os.Setenv("MACHINE_STORAGE_PATH", "")
}

func TestGetDockerDir(t *testing.T) {
	homeDir := mcnutils.GetHomeDir()
	baseDir := GetBaseDir()

	if strings.Index(baseDir, homeDir) != 0 {
		t.Fatalf("expected base dir with prefix %s; received %s", homeDir, baseDir)
	}
}

func TestGetMachineDir(t *testing.T) {
	root := "/tmp"
	os.Setenv("MACHINE_STORAGE_PATH", root)
	machineDir := GetMachineDir()

	if strings.Index(machineDir, root) != 0 {
		t.Fatalf("expected machine dir with prefix %s; received %s", root, machineDir)
	}

	path, filename := path.Split(machineDir)
	if strings.Index(path, root) != 0 {
		t.Fatalf("expected base path of %s; received %s", root, path)
	}
	if filename != "machines" {
		t.Fatalf("expected machine dir \"machines\"; received %s", filename)
	}
	os.Setenv("MACHINE_STORAGE_PATH", "")
}

func TestGetMachineCertDir(t *testing.T) {
	root := "/tmp"
	os.Setenv("MACHINE_STORAGE_PATH", root)
	clientDir := GetMachineCertDir()

	if strings.Index(clientDir, root) != 0 {
		t.Fatalf("expected machine client cert dir with prefix %s; received %s", root, clientDir)
	}

	path, filename := path.Split(clientDir)
	if strings.Index(path, root) != 0 {
		t.Fatalf("expected base path of %s; received %s", root, path)
	}
	if filename != "certs" {
		t.Fatalf("expected machine client dir \"certs\"; received %s", filename)
	}
	os.Setenv("MACHINE_STORAGE_PATH", "")
}
