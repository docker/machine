// +build !darwin,!linux,!windows

package vmwarefusion

import "github.com/docker/machine/libmachine/drivers"

func NewDriver(hostName, storePath string) drivers.Driver {
	return drivers.NewDriverNotSupported("vmwarefusion", hostName, storePath)
}

func DhcpConfigFiles() string {
	return ""
}

func DhcpLeaseFiles() string {
	return ""
}

func SetUmask() {
}

func setVmwareCmd(cmd string) string {
	return ""
}
