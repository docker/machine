package provision

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
)

func TestGenerateDockerConfigBoot2Docker(t *testing.T) {
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.FakeDriver{},
	}
	dockerPort := 1234
	authConfig := auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}
	engineConfigPath := "/var/lib/boot2docker/profile"

	dockerCfg, err := p.GenerateDockerConfig(dockerPort, authConfig)
	if err != nil {
		t.Fatal(err)
	}

	if dockerCfg.EngineConfigPath != engineConfigPath {
		t.Fatalf("expected engine path %s; received %s", engineConfigPath, dockerCfg.EngineConfigPath)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("-H tcp://0.0.0.0:%d", dockerPort)) == -1 {
		t.Fatalf("-H docker port invalid; expected %d", dockerPort)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("CACERT=%s", authConfig.CaCertRemotePath)) == -1 {
		t.Fatalf("CACERT option invalid; expected %s", authConfig.CaCertRemotePath)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("SERVERKEY=%s", authConfig.ServerKeyRemotePath)) == -1 {
		t.Fatalf("SERVERKEY option invalid; expected %s", authConfig.ServerKeyRemotePath)
	}

	if strings.Index(dockerCfg.EngineConfig, fmt.Sprintf("SERVERCERT=%s", authConfig.ServerCertRemotePath)) == -1 {
		t.Fatalf("SERVERCERT option invalid; expected %s", authConfig.ServerCertRemotePath)
	}
}

func TestMachinePortBoot2Docker(t *testing.T) {
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.FakeDriver{},
	}
	dockerPort := 2376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	authConfig := auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}

	cfg, err := p.GenerateDockerConfig(dockerPort, authConfig)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile("-H tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineConfig)
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
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.FakeDriver{},
	}
	dockerPort := 3376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	authConfig := auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}

	cfg, err := p.GenerateDockerConfig(dockerPort, authConfig)
	if err != nil {
		t.Fatal(err)
	}

	re := regexp.MustCompile("-H tcp://.*:(.+)")
	m := re.FindStringSubmatch(cfg.EngineConfig)
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
