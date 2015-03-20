package libmachine

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	_ "github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
)

const (
	hostTestName       = "test-host"
	hostTestDriverName = "none"
	hostTestCaCert     = "test-cert"
	hostTestPrivateKey = "test-key"
)

var (
	hostTestStorePath string
)

func getTestStore() (Store, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hostTestStorePath = tmpDir

	os.Setenv("MACHINE_STORAGE_PATH", tmpDir)

	return NewFilestore(tmpDir, hostTestCaCert, hostTestPrivateKey), nil
}

func cleanup() {
	os.RemoveAll(hostTestStorePath)
}

func getTestDriverFlags() *DriverOptionsMock {
	name := hostTestName
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"name":            name,
			"url":             "unix:///var/run/docker.sock",
			"swarm":           false,
			"swarm-host":      "",
			"swarm-master":    false,
			"swarm-discovery": "",
		},
	}
	return flags
}

func getDefaultTestHost() (*Host, error) {
	engineOptions := &engine.EngineOptions{}
	swarmOptions := &swarm.SwarmOptions{
		Master:    false,
		Host:      "",
		Discovery: "",
		Address:   "",
	}
	host, err := NewHost(hostTestName, hostTestDriverName, hostTestStorePath, hostTestCaCert, hostTestPrivateKey, engineOptions, swarmOptions)
	if err != nil {
		return nil, err
	}

	flags := getTestDriverFlags()
	if err := host.Driver.SetConfigFromFlags(flags); err != nil {
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

func TestGenerateDockerConfigNonLocal(t *testing.T) {
	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	dockerPort := 1234
	caCertPath := "/test/ca-cert"
	serverKeyPath := "/test/server-key"
	serverCertPath := "/test/server-cert"
	engineConfigPath := "/etc/default/docker"

	dockerCfg, err := host.generateDockerConfig(dockerPort, caCertPath, serverKeyPath, serverCertPath)
	if err != nil {
		t.Fatal(err)
	}

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

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err = store.Save(host); err != nil {
		t.Fatal(err)
	}

	host, err = store.Get(hostTestName)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := host.generateDockerConfig(dockerPort, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("--host=tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineConfig)
	if len(m) == 0 {
		t.Errorf("could not find port %d in engine config", dockerPort)
	}

	b := m[0]
	u := strings.Split(b, "=")
	url := u[1]
	url = strings.Replace(url, "'", "", -1)
	url = strings.Replace(url, "\\\"", "", -1)
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

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err = store.Save(host); err != nil {
		t.Fatal(err)
	}

	host, err = store.Get(hostTestName)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := host.generateDockerConfig(dockerPort, "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	re := regexp.MustCompile("--host=tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineConfig)
	if len(m) == 0 {
		t.Errorf("could not find port %d in engine config", dockerPort)
	}

	b := m[0]
	u := strings.Split(b, "=")
	url := u[1]
	url = strings.Replace(url, "'", "", -1)
	url = strings.Replace(url, "\\\"", "", -1)
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

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err = store.Save(host); err != nil {
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
