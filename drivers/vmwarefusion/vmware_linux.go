// +build linux

package vmwarefusion

import (
	"os/exec"
	"syscall"
)

func DhcpConfigFiles() string {
	return "/etc/vmware/vmnet*/dhcpd/dhcpd.conf"
}

func DhcpLeaseFiles() string {
	return "/etc/vmware/vmnet*/dhcpd/dhcpd.leases"
}

func SetUmask() {
	_ = syscall.Umask(022)
}

// detect the vmrun and vmware-vdiskmanager cmds' path if needed
func setVmwareCmd(cmd string) string {
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}
