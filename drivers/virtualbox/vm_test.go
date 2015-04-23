package virtualbox

import (
	"strings"
	"testing"
)

var (
	testVMInfoText = `
storagecontrollerbootable0="on"
memory=1024
cpus=2
"SATA-0-0"="/home/ehazlett/.boot2docker/boot2docker.iso"
"SATA-IsEjected"="off"
"SATA-1-0"="/home/ehazlett/vm/test/disk.vmdk"
"SATA-ImageUUID-1-0"="12345-abcdefg"
"SATA-2-0"="none"
nic1="nat"
`
)

func TestVMInfo(t *testing.T) {
	r := strings.NewReader(testVMInfoText)
	vm, err := parseVMInfo(r)
	if err != nil {
		t.Fatal(err)
	}

	vmCPUs := 2
	vmMemory := 1024
	if vm.CPUs != vmCPUs {
		t.Fatalf("expected %d cpus; received %d", vmCPUs, vm.CPUs)
	}

	if vm.Memory != vmMemory {
		t.Fatalf("expected memory %d; received %d", vmMemory, vm.Memory)
	}
}
