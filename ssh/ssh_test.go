package ssh_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	// "time"

	"github.com/docker/machine/ssh"
)

func TestGenerateSSHKey(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}

	filename := filepath.Join(tmpDir, "sshkey")

	if err := ssh.GenerateSSHKey(filename); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("expected ssh key at %s", filename)
	}

	// cleanup
	_ = os.RemoveAll(tmpDir)
}
