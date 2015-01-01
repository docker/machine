package virtualbox

import (
	"errors"
	"regexp"
)

var (
	reDiskUUID = regexp.MustCompile(`SATA-ImageUUID-1-0\"=\"(.*)\"`)
)

// Disk
type vmDisk struct {
	UUID string
}

func getDiskInfo(vmName string) (*vmDisk, error) {
	out, err := vbmOut("showvminfo", vmName, "--details", "--machinereadable")
	if err != nil {
		return nil, err
	}

	res := reDiskUUID.FindStringSubmatch(string(out))
	if res == nil {
		return nil, errors.New("failed to parse disk info")
	}

	return &vmDisk{UUID: res[1]}, err
}

func cloneBoot2DockerVmDisk(vmName string, destPath string) error {
	info, err := getDiskInfo(vmName)
	if err != nil {
		return err
	}

	if err := vbm("clonehd", info.UUID, destPath); err != nil {
		return err
	}

	return nil
}
