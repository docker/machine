package virtualbox

import (
	"testing"

	"github.com/docker/machine/drivers/vmwarevsphere/errors"
	"github.com/stretchr/testify/assert"
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
	vbox := &VBoxManagerMock{
		args:   "showvminfo default --machinereadable",
		stdOut: testDiskInfoText,
	}

	disk, err := getVMDiskInfo("default", vbox)

	assert.Equal(t, "/home/ehazlett/vm/test/disk.vmdk", disk.Path)
	assert.Equal(t, "12345-abcdefg", disk.UUID)
	assert.NoError(t, err)
}

func TestVMDiskInfoError(t *testing.T) {
	vbox := &VBoxManagerMock{
		args: "showvminfo default --machinereadable",
		err:  errors.New("BUG"),
	}

	disk, err := getVMDiskInfo("default", vbox)

	assert.Nil(t, disk)
	assert.EqualError(t, err, "BUG")
}

func TestVMDiskInfoInvalidOutput(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "showvminfo default --machinereadable",
		stdOut: "INVALID",
	}

	disk, err := getVMDiskInfo("default", vbox)

	assert.Empty(t, disk.Path)
	assert.Empty(t, disk.UUID)
	assert.NoError(t, err)
}
