package host

import (
	"path/filepath"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
)

type AuthOptionsV1 struct {
	StorePath            string
	CaCertPath           string
	CaCertRemotePath     string
	ServerCertPath       string
	ServerKeyPath        string
	ClientKeyPath        string
	ServerCertRemotePath string
	ServerKeyRemotePath  string
	PrivateKeyPath       string
	ClientCertPath       string
}

type HostOptionsV1 struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.EngineOptions
	SwarmOptions  *swarm.SwarmOptions
	AuthOptions   *AuthOptionsV1
}

type HostV1 struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *HostOptionsV1
	Name          string `json:"-"`
	StorePath     string
}

func MigrateHostV1ToHostV2(hostV1 *HostV1) *HostV2 {
	// Changed:  Put StorePath directly in AuthOptions (for provisioning),
	// and AuthOptions.PrivateKeyPath => AuthOptions.CaPrivateKeyPath
	// Also, CertDir has been added.

	globalStorePath := filepath.Dir(filepath.Dir(hostV1.StorePath))

	h := &HostV2{
		ConfigVersion: hostV1.ConfigVersion,
		Driver:        hostV1.Driver,
		Name:          hostV1.Driver.GetMachineName(),
		DriverName:    hostV1.DriverName,
		HostOptions: &HostOptions{
			Driver:        hostV1.HostOptions.Driver,
			Memory:        hostV1.HostOptions.Memory,
			Disk:          hostV1.HostOptions.Disk,
			EngineOptions: hostV1.HostOptions.EngineOptions,
			SwarmOptions:  hostV1.HostOptions.SwarmOptions,
			AuthOptions: &auth.AuthOptions{
				CertDir:              filepath.Join(globalStorePath, "certs"),
				CaCertPath:           hostV1.HostOptions.AuthOptions.CaCertPath,
				CaPrivateKeyPath:     hostV1.HostOptions.AuthOptions.PrivateKeyPath,
				CaCertRemotePath:     hostV1.HostOptions.AuthOptions.CaCertRemotePath,
				ServerCertPath:       hostV1.HostOptions.AuthOptions.ServerCertPath,
				ServerKeyPath:        hostV1.HostOptions.AuthOptions.ServerKeyPath,
				ClientKeyPath:        hostV1.HostOptions.AuthOptions.ClientKeyPath,
				ServerCertRemotePath: hostV1.HostOptions.AuthOptions.ServerCertRemotePath,
				ServerKeyRemotePath:  hostV1.HostOptions.AuthOptions.ServerKeyRemotePath,
				ClientCertPath:       hostV1.HostOptions.AuthOptions.ClientCertPath,
				StorePath:            globalStorePath,
			},
		},
	}

	return h
}
