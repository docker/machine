package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	_ "github.com/docker/machine/drivers/none"
)

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
