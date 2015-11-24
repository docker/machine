package virtualbox

import (
	"bytes"
	"io/ioutil"

	"github.com/docker/machine/libmachine/log"
)

// IsVTXDisabled checks if VT-X is disabled in the BIOS. If it is, the vm will fail to start.
// If we can't be sure it is disabled, we carry on and will check the vm logs after it's started.
// We want to check that either vmx or smd flags are present in /proc/cpuinfo.
func (d *Driver) IsVTXDisabled() bool {
	features := [2][]byte{
		{'v', 'm', 'x'},
		{'s', 'm', 'd'},
	}

	errmsg := "Couldn't check that VT-X/AMD-v is enabled. Will check that the vm is properly created: %v"
	content, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		log.Debugf(errmsg, err)
		return false
	}

	for _, v := range features {
		if bytes.Contains(content, v) {
			return false
		}
	}
	return true
}

func detectVBoxManageCmd() string {
	return detectVBoxManageCmdInPath()
}

func getShareDriveAndName() (string, string) {
	return "hosthome", "/home"
}
