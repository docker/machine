package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	_ "github.com/docker/machine/drivers/none"
	"github.com/docker/machine/utils"
)

const (
	hostTestName       = "test-host"
	hostTestDriverName = "none"
	hostTestStorePath  = "/test/path"
	hostTestCaCert     = "test-cert"
	hostTestPrivateKey = "test-key"
)

func getTestStore() (*Store, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Setenv("MACHINE_DIR", tmpDir)
	return NewStore(tmpDir, hostTestCaCert, hostTestPrivateKey), nil
}

func getTestDriverFlags() *DriverOptionsMock {
	name := hostTestName
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"name": name,
			"url":  "unix:///var/run/docker.sock",
		},
	}
	return flags
}

func getDefaultTestHost() (*Host, error) {
	host, err := NewHost(hostTestName, hostTestDriverName, hostTestStorePath, hostTestCaCert, hostTestPrivateKey)
	if err != nil {
		return nil, err
	}

	return host, nil
}

func TestLoadHostDoesNotExist(t *testing.T) {
	_, err := LoadHost("nope-not-here", "/nope/doesnotexist")
	if err == nil {
		t.Fatalf("expected error for non-existent host")
	}
}

func TestLoadHostExists(t *testing.T) {
	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}
	if host.Name != hostTestName {
		t.Fatalf("expected name %s; received %s", hostTestName, host.Name)
	}

	if host.DriverName != hostTestDriverName {
		t.Fatalf("expected driver %s; received %s", hostTestDriverName, host.DriverName)
	}

	if host.CaCertPath != hostTestCaCert {
		t.Fatalf("expected ca cert path %s; received %s", hostTestCaCert, host.CaCertPath)
	}

	if host.PrivateKeyPath != hostTestPrivateKey {
		t.Fatalf("expected key path %s; received %s", hostTestPrivateKey, host.PrivateKeyPath)
	}
}

func TestValidateHostnameValid(t *testing.T) {
	hosts := []string{
		"zomg",
		"test-ing",
		"some.h0st",
	}

	for _, v := range hosts {
		h, err := ValidateHostName(v)
		if err != nil {
			t.Fatal("Invalid hostname")
		}

		if h != v {
			t.Fatal("Hostname doesn't match")
		}
	}
}

func TestValidateHostnameInvalid(t *testing.T) {
	hosts := []string{
		"zom_g",
		"test$ing",
		"some😄host",
	}

	for _, v := range hosts {
		_, err := ValidateHostName(v)
		if err == nil {
			t.Fatal("No error returned")
		}
	}
}

func TestGenerateClientCertificate(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Setenv("MACHINE_DIR", tmpDir)

	caCertPath := filepath.Join(tmpDir, "ca.pem")
	caKeyPath := filepath.Join(tmpDir, "key.pem")
	testOrg := "test-org"
	bits := 2048
	if err := utils.GenerateCACertificate(caCertPath, caKeyPath, testOrg, bits); err != nil {
		t.Fatal(err)
	}

	if err := GenerateClientCertificate(caCertPath, caKeyPath); err != nil {
		t.Fatal(err)
	}

	clientCertPath := filepath.Join(utils.GetMachineDir(), "cert.pem")
	clientKeyPath := filepath.Join(utils.GetMachineDir(), "key.pem")
	if _, err := os.Stat(clientCertPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(clientKeyPath); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateDockerConfigNonLocal(t *testing.T) {
	host, err := NewHost(hostTestName, hostTestDriverName, hostTestStorePath, hostTestCaCert, hostTestPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	dockerPort := 1234
	caCertPath := "/test/ca-cert"
	serverKeyPath := "/test/server-key"
	serverCertPath := "/test/server-cert"
	engineConfigPath := "/etc/default/docker"

	dockerCfg := host.generateDockerConfig(dockerPort, caCertPath, serverKeyPath, serverCertPath)

	if dockerCfg.EngineConfigPath != engineConfigPath {
		t.Fatalf("expected engine path %s; received %s", engineConfigPath, dockerCfg.EngineConfigPath)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("--host=tcp://0.0.0.0:%d", dockerPort)) == -1 {
		t.Fatalf("--host docker port invalid; expected %d", dockerPort)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("--tlscacert=%s", caCertPath)) == -1 {
		t.Fatalf("--tlscacert option invalid; expected %s", caCertPath)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("--tlskey=%s", serverKeyPath)) == -1 {
		t.Fatalf("--tlskey option invalid; expected %s", serverKeyPath)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("--tlscert=%s", serverCertPath)) == -1 {
		t.Fatalf("--tlscert option invalid; expected %s", serverCertPath)
	}
}

func TestMachinePort(t *testing.T) {
	dockerPort := 2376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}
	flags := getTestDriverFlags()

	_, err = store.Create(hostTestName, hostTestDriverName, flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err := store.Load(hostTestName)
	if err != nil {
		t.Fatal(err)
	}
	cfg := host.generateDockerConfig(dockerPort, "", "", "")

	re := regexp.MustCompile("--host=tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineConfig)
	if len(m) == 0 {
		t.Errorf("could not find port %d in engine config", dockerPort)
	}

	b := m[0]
	u := strings.Split(b, "=")
	url := u[1]
	url = strings.Replace(url, "'", "", -1)
	if url != bindUrl {
		t.Errorf("expected url %s; received %s", bindUrl, url)
	}

	if err := store.Remove(hostTestName, true); err != nil {
		t.Fatal(err)
	}
}

func TestMachineCustomPort(t *testing.T) {
	dockerPort := 3376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}
	flags := getTestDriverFlags()

	_, err = store.Create(hostTestName, hostTestDriverName, flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err := store.Load(hostTestName)
	if err != nil {
		t.Fatal(err)
	}
	cfg := host.generateDockerConfig(dockerPort, "", "", "")

	re := regexp.MustCompile("--host=tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineConfig)
	if len(m) == 0 {
		t.Errorf("could not find port %d in engine config", dockerPort)
	}

	b := m[0]
	u := strings.Split(b, "=")
	url := u[1]
	url = strings.Replace(url, "'", "", -1)
	if url != bindUrl {
		t.Errorf("expected url %s; received %s", bindUrl, url)
	}

	if err := store.Remove(hostTestName, true); err != nil {
		t.Fatal(err)
	}
}

func TestHostConfig(t *testing.T) {
	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	flags := getTestDriverFlags()
	host, err := store.Create(hostTestName, hostTestDriverName, flags)
	if err != nil {
		t.Fatal(err)
	}

	if err := host.SaveConfig(); err != nil {
		t.Fatal(err)
	}

	if err := host.LoadConfig(); err != nil {
		t.Fatal(err)
	}

	// cleanup
	if err := store.Remove(hostTestName, true); err != nil {
		t.Fatal(err)
	}
}
