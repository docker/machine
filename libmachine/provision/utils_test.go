package provision

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
)

func TestGenerateDockerOptionsBoot2Docker(t *testing.T) {
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.FakeDriver{},
	}
	dockerPort := 1234
	authOptions := auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}
	engineConfigPath := "/var/lib/boot2docker/profile"

	dockerCfg, err := p.GenerateDockerOptions(dockerPort, authOptions)
	if err != nil {
		t.Fatal(err)
	}

	if dockerCfg.EngineOptionsPath != engineConfigPath {
		t.Fatalf("expected engine path %s; received %s", engineConfigPath, dockerCfg.EngineOptionsPath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("-H tcp://0.0.0.0:%d", dockerPort)) == -1 {
		t.Fatalf("-H docker port invalid; expected %d", dockerPort)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("CACERT=%s", authOptions.CaCertRemotePath)) == -1 {
		t.Fatalf("CACERT option invalid; expected %s", authOptions.CaCertRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("SERVERKEY=%s", authOptions.ServerKeyRemotePath)) == -1 {
		t.Fatalf("SERVERKEY option invalid; expected %s", authOptions.ServerKeyRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("SERVERCERT=%s", authOptions.ServerCertRemotePath)) == -1 {
		t.Fatalf("SERVERCERT option invalid; expected %s", authOptions.ServerCertRemotePath)
	}
}

func TestMachinePortBoot2Docker(t *testing.T) {
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.FakeDriver{},
	}
	dockerPort := 2376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	authOptions := auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}

	cfg, err := p.GenerateDockerOptions(dockerPort, authOptions)
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
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.FakeDriver{},
	}
	dockerPort := 3376
	bindUrl := fmt.Sprintf("tcp://0.0.0.0:%d", dockerPort)
	authOptions := auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}

	cfg, err := p.GenerateDockerOptions(dockerPort, authOptions)
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
