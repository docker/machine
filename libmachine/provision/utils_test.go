package provision

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
)

func engineOptions() engine.EngineOptions {
	return engine.EngineOptions{
		StorageDriver: "aufs",
	}
}

func TestGenerateDockerOptionsBoot2Docker(t *testing.T) {
	g := GenericProvisioner{
		Driver:        &fakedriver.FakeDriver{},
		EngineOptions: engineOptions(),
	}
	p := &Boot2DockerProvisioner{
		DebianProvisioner{
			g,
		},
	}
	dockerPort := 1234
	p.AuthOptions = auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}

	dockerCfg, err := p.GenerateDockerOptions(dockerPort)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("-H tcp://0.0.0.0:%d", dockerPort)) == -1 {
		t.Fatalf("-H docker port invalid; expected %d", dockerPort)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("--tlscacert %s", p.AuthOptions.CaCertRemotePath)) == -1 {
		t.Fatalf("--tlscacert option invalid; expected %s", p.AuthOptions.CaCertRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("--tlscert %s", p.AuthOptions.ServerCertRemotePath)) == -1 {
		t.Fatalf("--tlscert option invalid; expected %s", p.AuthOptions.ServerCertRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("--tlskey %s", p.AuthOptions.ServerKeyRemotePath)) == -1 {
		t.Fatalf("--tlskey option invalid; expected %s", p.AuthOptions.ServerKeyRemotePath)
	}
}

func TestMachinePortBoot2Docker(t *testing.T) {
	g := GenericProvisioner{
		Driver:        &fakedriver.FakeDriver{},
		EngineOptions: engineOptions(),
	}
	p := &Boot2DockerProvisioner{
		DebianProvisioner{
			g,
		},
	}
	dockerPort := 2376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	p.AuthOptions = auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}

	cfg, err := p.GenerateDockerOptions(dockerPort)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("-H tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineOptions)
	if len(m) == 0 {
		t.Errorf("could not find port %d in engine config", dockerPort)
	}

	b := m[0]
	u := strings.Split(b, " ")
	url := u[1]
	url = strings.Replace(url, "'", "", -1)
	url = strings.Replace(url, "\\\"", "", -1)
	if url != bindUrl {
		t.Errorf("expected url %s; received %s", bindUrl, url)
	}
}

func TestMachineCustomPortBoot2Docker(t *testing.T) {
	g := GenericProvisioner{
		Driver:        &fakedriver.FakeDriver{},
		EngineOptions: engineOptions(),
	}
	p := &Boot2DockerProvisioner{
		DebianProvisioner{
			g,
		},
	}
	dockerPort := 3376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	p.AuthOptions = auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}

	cfg, err := p.GenerateDockerOptions(dockerPort)
	if err != nil {
		t.Fatal(err)
	}

	re := regexp.MustCompile("-H tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineOptions)
	if len(m) == 0 {
		t.Errorf("could not find port %d in engine config", dockerPort)
	}

	b := m[0]
	u := strings.Split(b, " ")
	url := u[1]
	url = strings.Replace(url, "'", "", -1)
	url = strings.Replace(url, "\\\"", "", -1)
	if url != bindUrl {
		t.Errorf("expected url %s; received %s", bindUrl, url)
	}
}
