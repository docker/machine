package machine

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestMachineLoadMachine(t *testing.T) {
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}

	st := NewStore(TestStoreDir, "", "")

	m, err := getTestMachine()
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Save(m); err != nil {
		t.Fatal(err)
	}

	machinePath := filepath.Join(st.Path, "test")

	mcn, err := loadMachine(m.Name, machinePath)
	if err != nil {
		t.Fatal(err)
	}

	if mcn.Name != m.Name {
		t.Fatalf("expected machine name %s; received %s", m.Name, mcn.Name)
	}

	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestMachineValidHostname(t *testing.T) {
	hostname := "test-host"
	name, err := ValidateHostname(hostname)
	if err != nil {
		t.Fatal(err)
	}

	if name != hostname {
		t.Fatal("expected hostname %s; received %s", hostname, name)
	}
}

func TestMachineValidHostnameInvalid(t *testing.T) {
	hostname := "test,%"
	if _, err := ValidateHostname(hostname); err == nil {
		t.Fatal("expected failure")
	}
}

func TestMachineGenerateDockerConfig(t *testing.T) {
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}

	st := NewStore(TestStoreDir, "", "")

	m, err := getTestMachine()
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Save(m); err != nil {
		t.Fatal(err)
	}

	port := 1234
	caCertPath := "/tmp/ca"
	serverKeyPath := "/tmp/server-key"
	serverCertPath := "/tmp/server-cert"

	config := m.generateDockerConfig(port, caCertPath, serverKeyPath, serverCertPath)

	if !strings.Contains(config.EngineConfig, caCertPath) {
		t.Fatal("failed to find caCertPath")
	}

	if !strings.Contains(config.EngineConfig, serverKeyPath) {
		t.Fatal("failed to find serverKeyPath")
	}

	if !strings.Contains(config.EngineConfig, serverCertPath) {
		t.Fatal("failed to find serverCertPath")
	}

	if !strings.Contains(config.EngineConfig, fmt.Sprintf("-H tcp://0.0.0.0:%d", port)) {
		t.Fatal("failed to find serverCertPath")
	}

	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
}
