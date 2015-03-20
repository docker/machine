package libmachine

import (
	"fmt"
	"io/ioutil"
	"os"

	"testing"

	_ "github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/auth"
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
	hostOptions := &HostOptions{
		EngineConfig: &engine.EngineOptions{},
		SwarmConfig: &swarm.SwarmOptions{
			Master:    false,
			Host:      "",
			Discovery: "",
			Address:   "",
		},
		AuthConfig: &auth.AuthOptions{
			CaCertPath:     hostTestCaCert,
			PrivateKeyPath: hostTestPrivateKey,
		},
	}
	host, err := NewHost(hostTestName, hostTestDriverName, hostOptions)
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
	authConfig := host.HostConfig.AuthConfig
	if host.Name != hostTestName {
		t.Fatalf("expected name %s; received %s", hostTestName, host.Name)
	}

	if host.DriverName != hostTestDriverName {
		t.Fatalf("expected driver %s; received %s", hostTestDriverName, host.DriverName)
	}

	if authConfig.CaCertPath != hostTestCaCert {
		t.Fatalf("expected ca cert path %s; received %s", hostTestCaCert, authConfig.CaCertPath)
	}

	if authConfig.PrivateKeyPath != hostTestPrivateKey {
		t.Fatalf("expected key path %s; received %s", hostTestPrivateKey, authConfig.PrivateKeyPath)
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
