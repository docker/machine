package virtualbox

import (
	"strings"
	"testing"
)

var (
	testDiskInfoText = `
storagecontrollerbootable0="on"
"SATA-0-0"="/home/ehazlett/.boot2docker/boot2docker.iso"
"SATA-IsEjected"="off"
"SATA-1-0"="/home/ehazlett/vm/test/disk.vmdk"
"SATA-ImageUUID-1-0"="12345-abcdefg"
"SATA-2-0"="none"
nic1="nat"
    `
)

func TestVMDiskInfo(t *testing.T) {
	r := strings.NewReader(testDiskInfoText)
	disk, err := parseDiskInfo(r)
	if err != nil {
		t.Fatal(err)
	}

	diskPath := "/home/ehazlett/vm/test/disk.vmdk"
	diskUUID := "12345-abcdefg"
	if disk.Path != diskPath {
		t.Fatalf("expected disk path %s", diskPath)
	}

	if disk.UUID != diskUUID {
		t.Fatalf("expected disk uuid %s", diskUUID)
	}
}
