// +build darwin

package vmwarefusion

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func DhcpConfigFiles() string {
	return "/Library/Preferences/VMware Fusion/vmnet*/dhcpd.conf"
}

func DhcpLeaseFiles() string {
	return "/var/db/vmware/*.leases"
}

func SetUmask() {
	_ = syscall.Umask(022)
}

// detect the vmrun and vmware-vdiskmanager cmds' path if needed
func setVmwareCmd(cmd string) string {
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	for _, fp := range []string{
		"/Applications/VMware Fusion.app/Contents/Library/",
	} {
		p := filepath.Join(fp, cmd)
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}
	return cmd
}
