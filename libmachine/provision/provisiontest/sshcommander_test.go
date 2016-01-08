package provisiontest

import "testing"

func TestStatSSHCommand(t *testing.T) {
	sshCmder := FakeSSHCommander{FilesystemType: "btrfs"}
	output, err := sshCmder.SSHCommand("stat -f -c %T /var/lib")
	if err != nil || output != "btrfs\n" {
		t.Fatal("FakeSSHCommander should have returned btrfs and no error but returned '", output, "' and error", err)
	}
}
