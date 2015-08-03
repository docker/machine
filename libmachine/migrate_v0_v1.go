package libmachine

import (
	"path/filepath"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/utils"
)

// In the 0.0.1 => 0.0.2 transition, the JSON representation of
// machines changed from a "flat" to a more "nested" structure
// for various options and configuration settings.  To preserve
// compatibility with existing machines, these migration functions
// have been introduced.  They preserve backwards compat at the expense
// of some duplicated information.

// validates host config and modifies if needed
// this is used for configuration updates
func MigrateHostV0ToHostV1(hostV0 *HostV0) *Host {
	host := &Host{}

	certInfoV0 := getCertInfoFromHost(hostV0)

	host.HostOptions = &HostOptions{}
	host.HostOptions.EngineOptions = &engine.EngineOptions{}
	host.HostOptions.SwarmOptions = &swarm.SwarmOptions{
		Address:   "",
		Discovery: hostV0.SwarmDiscovery,
		Host:      hostV0.SwarmHost,
		Master:    hostV0.SwarmMaster,
	}
	host.HostOptions.AuthOptions = &auth.AuthOptions{
		StorePath:            hostV0.StorePath,
		CaCertPath:           certInfoV0.CaCertPath,
		CaCertRemotePath:     "",
		ServerCertPath:       certInfoV0.ServerCertPath,
		ServerKeyPath:        certInfoV0.ServerKeyPath,
		ClientKeyPath:        certInfoV0.ClientKeyPath,
		ServerCertRemotePath: "",
		ServerKeyRemotePath:  "",
		PrivateKeyPath:       certInfoV0.CaKeyPath,
		ClientCertPath:       certInfoV0.ClientCertPath,
	}

	return host
}

// fills nested host metadata and modifies if needed
// this is used for configuration updates
func MigrateHostMetadataV0ToHostMetadataV1(m *HostMetadataV0) *HostMetadata {
	hostMetadata := &HostMetadata{}
	hostMetadata.DriverName = m.DriverName
	hostMetadata.HostOptions.EngineOptions = &engine.EngineOptions{}
	hostMetadata.HostOptions.AuthOptions = &auth.AuthOptions{
		StorePath:            m.StorePath,
		CaCertPath:           m.CaCertPath,
		CaCertRemotePath:     "",
		ServerCertPath:       m.ServerCertPath,
		ServerKeyPath:        m.ServerKeyPath,
		ClientKeyPath:        "",
		ServerCertRemotePath: "",
		ServerKeyRemotePath:  "",
		PrivateKeyPath:       m.PrivateKeyPath,
		ClientCertPath:       m.ClientCertPath,
	}

	return hostMetadata
}

func getCertInfoFromHost(h *HostV0) CertPathInfo {
	// setup cert paths
	caCertPath := h.CaCertPath
	caKeyPath := h.PrivateKeyPath
	clientCertPath := h.ClientCertPath
	clientKeyPath := h.ClientKeyPath
	serverCertPath := h.ServerCertPath
	serverKeyPath := h.ServerKeyPath

	if caCertPath == "" {
		caCertPath = filepath.Join(utils.GetMachineCertDir(), "ca.pem")
	}

	if caKeyPath == "" {
		caKeyPath = filepath.Join(utils.GetMachineCertDir(), "ca-key.pem")
	}

	if clientCertPath == "" {
		clientCertPath = filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	}

	if clientKeyPath == "" {
		clientKeyPath = filepath.Join(utils.GetMachineCertDir(), "key.pem")
	}

	if serverCertPath == "" {
		serverCertPath = filepath.Join(utils.GetMachineCertDir(), "server.pem")
	}

	if serverKeyPath == "" {
		serverKeyPath = filepath.Join(utils.GetMachineCertDir(), "server-key.pem")
	}

	return CertPathInfo{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		ServerCertPath: serverCertPath,
		ServerKeyPath:  serverKeyPath,
	}
}
