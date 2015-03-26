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
func FillNestedHost(host *Host) *Host {
	certInfo := getCertInfoFromHost(host)

	if host.HostOptions == nil {
		host.HostOptions = &HostOptions{}
	}
	if host.HostOptions.EngineOptions == nil {
		host.HostOptions.EngineOptions = &engine.EngineOptions{}
	}

	if host.HostOptions.SwarmOptions == nil {
		host.HostOptions.SwarmOptions = &swarm.SwarmOptions{
			Address:   "",
			Discovery: host.SwarmDiscovery,
			Host:      host.SwarmHost,
			Master:    host.SwarmMaster,
		}
	}

	host.HostOptions.AuthOptions = &auth.AuthOptions{
		StorePath:            host.StorePath,
		CaCertPath:           certInfo.CaCertPath,
		CaCertRemotePath:     "",
		ServerCertPath:       certInfo.ServerCertPath,
		ServerKeyPath:        certInfo.ServerKeyPath,
		ClientKeyPath:        certInfo.ClientKeyPath,
		ServerCertRemotePath: "",
		ServerKeyRemotePath:  "",
		PrivateKeyPath:       certInfo.CaKeyPath,
		ClientCertPath:       certInfo.ClientCertPath,
	}

	return host
}

// fills nested host metadata and modifies if needed
// this is used for configuration updates
func FillNestedHostMetadata(m *HostMetadata) *HostMetadata {
	if m.HostOptions.EngineOptions == nil {
		m.HostOptions.EngineOptions = &engine.EngineOptions{}
	}

	if m.HostOptions.AuthOptions == nil {
		m.HostOptions.AuthOptions = &auth.AuthOptions{
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
	}

	return m
}

func getCertInfoFromHost(h *Host) CertPathInfo {
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
