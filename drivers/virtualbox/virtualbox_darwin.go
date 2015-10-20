package virtualbox

import (
	"strings"
	"syscall"

	"github.com/docker/machine/libmachine/log"
)

// IsVTXDisabled checks if VT-X is disabled in the BIOS. If it is, the vm will fail to start.
// If we can't be sure it is disabled, we carry on and will check the vm logs after it's started.
func (d *Driver) IsVTXDisabled() bool {
	errmsg := "Couldn't check that VT-X/AMD-v is enabled. Will check that the vm is properly created: %v"
	features, err := syscall.Sysctl("machdep.cpu.features")
	if err != nil {
		log.Debugf(errmsg, err)
		return false
	}

	disabled := !strings.Contains(features, "VMX")
	return disabled
}
