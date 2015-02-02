package utils

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetBaseDir(t *testing.T) {
	// reset any override env var
	homeDir := GetHomeDir()
	baseDir := GetBaseDir()

	if strings.Index(homeDir, baseDir) != 0 {
		t.Fatalf("expected base dir with prefix %s; received %s", homeDir, baseDir)
	}
}

func TestGetCustomBaseDir(t *testing.T) {
	root := "/tmp"
	os.Setenv("MACHINE_DIR", root)
	baseDir := GetBaseDir()

	if strings.Index(root, baseDir) != 0 {
		t.Fatalf("expected base dir with prefix %s; received %s", root, baseDir)
	}
	os.Setenv("MACHINE_DIR", "")
}

func TestGetDockerDir(t *testing.T) {
	root := "/tmp"
	os.Setenv("MACHINE_DIR", root)
	dockerDir := GetDockerDir()

	if strings.Index(dockerDir, root) != 0 {
		t.Fatalf("expected docker dir with prefix %s; received %s", root, dockerDir)
	}

	path, filename := path.Split(dockerDir)
	if strings.Index(path, root) != 0 {
		t.Fatalf("expected base path of %s; received %s", root, path)
	}
	if filename != ".docker" {
		t.Fatalf("expected docker dir \".docker\"; received %s", filename)
	}
	os.Setenv("MACHINE_DIR", "")
}

func TestGetMachineDir(t *testing.T) {
	root := "/tmp"
	os.Setenv("MACHINE_DIR", root)
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
	os.Setenv("MACHINE_DIR", "")
}

func TestGetMachineClientCertDir(t *testing.T) {
	root := "/tmp"
	os.Setenv("MACHINE_DIR", root)
	clientDir := GetMachineClientCertDir()

	if strings.Index(clientDir, root) != 0 {
		t.Fatalf("expected machine client cert dir with prefix %s; received %s", root, clientDir)
	}

	path, filename := path.Split(clientDir)
	if strings.Index(path, root) != 0 {
		t.Fatalf("expected base path of %s; received %s", root, path)
	}
	if filename != ".client" {
		t.Fatalf("expected machine client dir \".client\"; received %s", filename)
	}
	os.Setenv("MACHINE_DIR", "")
}

func TestCopyFile(t *testing.T) {
	testStr := "test-machine"

	srcFile, err := ioutil.TempFile("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}
	srcFi, err := srcFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	srcFile.Write([]byte(testStr))
	srcFile.Close()

	srcFilePath := filepath.Join(os.TempDir(), srcFi.Name())

	destFile, err := ioutil.TempFile("", "machine-copy-test-")
	if err != nil {
		t.Fatal(err)
	}

	destFi, err := destFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	destFile.Close()

	destFilePath := filepath.Join(os.TempDir(), destFi.Name())

	if err := CopyFile(srcFilePath, destFilePath); err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile(destFilePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != testStr {
		t.Fatalf("expected data \"%s\"; received \"%\"", testStr, string(data))
	}
}

func TestGetUsername(t *testing.T) {
	currentUser := "unknown"
	switch runtime.GOOS {
	case "darwin", "linux":
		currentUser = os.Getenv("USER")
	case "windows":
		currentUser = os.Getenv("USERNAME")
	}

	username := GetUsername()
	if username != currentUser {
		t.Fatalf("expected username %s; received %s", currentUser, username)
	}
}
