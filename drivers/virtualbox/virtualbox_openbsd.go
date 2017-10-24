package virtualbox

import "github.com/docker/machine/libmachine/drivers"

func detectVBoxManageCmd() string {
	return ""
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return drivers.NewDriverNotSupported("virtualbox", hostName, storePath)
}
