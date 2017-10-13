// +build windows

package vmwarefusion

import (
	"os"
	"os/exec"
	"path/filepath"
)

func DhcpConfigFiles() string {
	return `C:\ProgramData\VMware\vmnetdhcp.conf`
}

func DhcpLeaseFiles() string {
	return `C:\ProgramData\VMware\vmnetdhcp.leases`
}

func SetUmask() {
}

func setVmwareCmd(cmd string) string {
	cmd = cmd + ".exe"

	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	for _, fp := range []string{
		`C:\Program Files (x86)\VMware\VMware Workstation`,
		`C:\Program Files\VMware\VMware Workstation`,
	} {
		p := filepath.Join(fp, cmd)
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}
	return cmd
}
