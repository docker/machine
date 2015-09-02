package host

import "github.com/docker/machine/libmachine/drivers"

type HostV2 struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *HostOptions
	Name          string
}
