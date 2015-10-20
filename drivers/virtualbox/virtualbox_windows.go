package virtualbox

import (
	"strings"

	"github.com/docker/machine/libmachine/log"
)

// IsVTXDisabled checks if VT-X is disabled in the BIOS. If it is, the vm will fail to start.
// If we can't be sure it is disabled, we carry on and will check the vm logs after it's started.
func (d *Driver) IsVTXDisabled() bool {
	errmsg := "Couldn't check that VT-X/AMD-v is enabled. Will check that the vm is properly created: %v"
	output, err := cmdOutput("wmic", "cpu", "get", "VirtualizationFirmwareEnabled")
	if err != nil {
		log.Debugf(errmsg, err)
		return false
	}

	disabled := strings.Contains(output, "FALSE")
	return disabled
}
