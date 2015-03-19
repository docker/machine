package libmachine

import (
	"github.com/docker/machine/drivers"
)

type EngineOptions struct {
	Dns            []string
	GraphDir       string
	Ipv6           bool
	Labels         []string
	LogLevel       string
	StorageDriver  string
	SelinuxEnabled bool
	TlsCaCert      string
	TlsCert        string
	TlsKey         string
	TlsVerify      bool
	RegistryMirror []string
}

type SwarmOptions struct {
	Address    string
	Discovery  string
	Master     bool
	Host       string
	Strategy   string
	Heartbeat  int
	Overcommit float64
	TlsCaCert  string
	TlsCert    string
	TlsKey     string
	TlsVerify  bool
}

type HostOptions struct {
	Driver        string
	Memory        int
	Disk          int
	DriverOptions drivers.DriverOptions
	EngineOptions *EngineOptions
	SwarmOptions  *SwarmOptions
}
