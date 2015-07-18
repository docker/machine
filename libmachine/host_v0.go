package libmachine

import "github.com/docker/machine/drivers"

type HostV0 struct {
	Name          string `json:"-"`
	Driver        drivers.Driver
	DriverName    string
	ConfigVersion int
	HostOptions   *HostOptions

	StorePath      string
	CaCertPath     string
	PrivateKeyPath string
	ServerCertPath string
	ServerKeyPath  string
	ClientCertPath string
	SwarmHost      string
	SwarmMaster    bool
	SwarmDiscovery string
	ClientKeyPath  string
}

type HostMetadataV0 struct {
	HostOptions HostOptions
	DriverName  string

	StorePath      string
	CaCertPath     string
	PrivateKeyPath string
	ServerCertPath string
	ServerKeyPath  string
	ClientCertPath string
}
