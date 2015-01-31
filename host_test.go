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

func TestLoadHostDoesNotExist(t *testing.T) {
	_, err := LoadHost("nope-not-here", "/nope/doesnotexist")
	if err == nil {
		t.Fatal("Expected error for non-existent host")
	}
}

func TestLoadHostExists(t *testing.T) {
	name := "test-host"
	driver := "none"
	storePath := "/test/path"
	caCert := "test-cert"
	privateKey := "test-key"
	host, err := NewHost(name, driver, storePath, caCert, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	if host.Name != name {
		t.Fatal("expected name %s received %s", name, host.Name)
	}

	if host.DriverName != driver {
		t.Fatal("expected driver %s received %s", driver, host.DriverName)
	}

	if host.CaCertPath != caCert {
		t.Fatal("expected ca cert path %s received %s", caCert, host.CaCertPath)
	}

	if host.PrivateKeyPath != privateKey {
		t.Fatal("expected key path %s received %s", privateKey, host.PrivateKeyPath)
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
		t.Fatal(err)
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

	// cleanup
	_ = os.RemoveAll(tmpDir)
}

func TestMachinePort(t *testing.T) {
	dockerPort := 2376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": bindUrl,
		},
	}

	store := NewStore("", "", "")

	_, err := store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err := store.Load("test")
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

	if err := store.Remove("test", true); err != nil {
		t.Fatal(err)
	}
}

func TestMachineCustomPort(t *testing.T) {
	dockerPort := 3376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": bindUrl,
		},
	}

	store := NewStore("", "", "")

	_, err := store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err := store.Load("test")
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

	if err := store.Remove("test", true); err != nil {
		t.Fatal(err)
	}
}
