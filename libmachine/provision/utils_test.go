package provision

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/state"
)

type FakeDriver struct{}

func (d FakeDriver) DriverName() string {
	return "fakedriver"
}

func (d FakeDriver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d FakeDriver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d FakeDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d FakeDriver) GetURL() (string, error) {
	return "", nil
}

func (d FakeDriver) GetMachineName() string {
	return ""
}

func (d FakeDriver) GetProviderType() provider.ProviderType {
	return provider.None
}

func (d FakeDriver) GetIP() (string, error) {
	return "1.2.3.4", nil
}

func (d FakeDriver) GetSSHHostname() (string, error) {
	return "", nil
}

func (d FakeDriver) GetSSHKeyPath() string {
	return ""
}

func (d FakeDriver) GetSSHPort() (int, error) {
	return 0, nil
}

func (d FakeDriver) GetSSHUsername() string {
	return ""
}

func (d FakeDriver) GetState() (state.State, error) {
	return state.Running, nil
}

func (d FakeDriver) PreCreateCheck() error {
	return nil
}

func (d FakeDriver) Create() error {
	return nil
}

func (d FakeDriver) Remove() error {
	return nil
}

func (d FakeDriver) Start() error {
	return nil
}

func (d FakeDriver) Stop() error {
	return nil
}

func (d FakeDriver) Restart() error {
	return nil
}

func (d FakeDriver) Kill() error {
	return nil
}

func (d FakeDriver) Upgrade() error {
	return nil
}

func (d FakeDriver) StartDocker() error {
	return nil
}

func (d FakeDriver) StopDocker() error {
	return nil
}

func (d FakeDriver) GetDockerConfigDir() string {
	return ""
}

func (d FakeDriver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return &exec.Cmd{}, nil
}

func TestGenerateDockerConfigBoot2Docker(t *testing.T) {
	p := &Boot2DockerProvisioner{
		Driver: FakeDriver{},
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
		Driver: FakeDriver{},
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
		Driver: FakeDriver{},
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
