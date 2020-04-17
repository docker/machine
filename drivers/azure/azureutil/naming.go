package azureutil

import (
	"fmt"
)

const (
	fmtNIC             = "%s-nic"
	fmtIP              = "%s-ip"
	fmtNSG             = "%s-firewall"
	fmtVM              = "%s"
	fmtOSDisk          = "%s-os-disk"
	fmtOSDiskContainer = "vhd-%s" // place vhds of VMs in separate containers for ease of cleanup
	fmtOSDiskBlob      = "%s-os-disk.vhd"
)

// ResourceNaming provides methods to construct Azure resource names for a given
// machine name.
type ResourceNaming string

// IP returns the Azure resource name for an IP address
func (r ResourceNaming) IP() string { return fmt.Sprintf(fmtIP, r) }

// NIC returns the Azure resource name for a network interface
func (r ResourceNaming) NIC() string { return fmt.Sprintf(fmtNIC, r) }

// NSG returns the Azure resource name for a network security group
func (r ResourceNaming) NSG() string { return fmt.Sprintf(fmtNSG, r) }

// VM returns the Azure resource name for a VM
func (r ResourceNaming) VM() string { return fmt.Sprintf(fmtVM, r) }

// OSDisk returns the Azure resource name for an OS disk
func (r ResourceNaming) OSDisk() string { return fmt.Sprintf(fmtOSDisk, r) }

// OSDiskContainer returns the Azure resource name for an OS disk container
func (r ResourceNaming) OSDiskContainer() string { return fmt.Sprintf(fmtOSDiskContainer, r) }

// OSDiskBlob returns the Azure resource name for an OS disk blob
func (r ResourceNaming) OSDiskBlob() string { return fmt.Sprintf(fmtOSDiskBlob, r) }
