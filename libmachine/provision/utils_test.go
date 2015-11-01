package provision

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
)

var (
	reDaemonListening = ":2376.*LISTEN"
)

func TestMatchNetstatOutMissing(t *testing.T) {
	nsOut := `Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:ssh             0.0.0.0:*               LISTEN
tcp        0     72 192.168.25.141:ssh      192.168.25.1:63213      ESTABLISHED
tcp        0      0 :::ssh                  :::*                    LISTEN
Active UNIX domain sockets (servers and established)
Proto RefCnt Flags       Type       State         I-Node Path
unix  2      [ ACC ]     STREAM     LISTENING      17990 /var/run/acpid.socket
unix  2      [ ACC ]     SEQPACKET  LISTENING      14233 /run/udev/control
unix  3      [ ]         STREAM     CONNECTED      18688
unix  3      [ ]         DGRAM                     14243
unix  3      [ ]         STREAM     CONNECTED      18689
unix  3      [ ]         DGRAM                     14242`
	if matchNetstatOut(reDaemonListening, nsOut) {
		t.Fatal("Expected not to match the netstat output as showing the daemon listening but got a match")
	}
}

func TestMatchNetstatOutPresent(t *testing.T) {
	nsOut := `Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:ssh             0.0.0.0:*               LISTEN
tcp        0     72 192.168.25.141:ssh      192.168.25.1:63235      ESTABLISHED
tcp        0      0 :::2376                 :::*                    LISTEN
tcp        0      0 :::ssh                  :::*                    LISTEN
Active UNIX domain sockets (servers and established)
Proto RefCnt Flags       Type       State         I-Node Path
unix  2      [ ACC ]     STREAM     LISTENING      17990 /var/run/acpid.socket
unix  2      [ ACC ]     SEQPACKET  LISTENING      14233 /run/udev/control
unix  2      [ ACC ]     STREAM     LISTENING      19365 /var/run/docker.sock
unix  3      [ ]         STREAM     CONNECTED      19774
unix  3      [ ]         STREAM     CONNECTED      19775
unix  3      [ ]         DGRAM                     14243
unix  3      [ ]         DGRAM                     14242`
	if !matchNetstatOut(reDaemonListening, nsOut) {
		t.Fatal("Expected to match the netstat output as showing the daemon listening but didn't")
	}
}

func TestGenerateDockerOptionsBoot2Docker(t *testing.T) {
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.Driver{},
	}
	dockerPort := 1234
	p.AuthOptions = auth.AuthOptions{
		CaCertRemotePath:     "/test/ca-cert",
		ServerKeyRemotePath:  "/test/server-key",
		ServerCertRemotePath: "/test/server-cert",
	}
	engineConfigPath := "/var/lib/boot2docker/profile"

	dockerCfg, err := p.GenerateDockerOptions(dockerPort)
	if err != nil {
		t.Fatal(err)
	}

	if dockerCfg.EngineOptionsPath != engineConfigPath {
		t.Fatalf("expected engine path %s; received %s", engineConfigPath, dockerCfg.EngineOptionsPath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("-H tcp://0.0.0.0:%d", dockerPort)) == -1 {
		t.Fatalf("-H docker port invalid; expected %d", dockerPort)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("CACERT=%s", p.AuthOptions.CaCertRemotePath)) == -1 {
		t.Fatalf("CACERT option invalid; expected %s", p.AuthOptions.CaCertRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("SERVERKEY=%s", p.AuthOptions.ServerKeyRemotePath)) == -1 {
		t.Fatalf("SERVERKEY option invalid; expected %s", p.AuthOptions.ServerKeyRemotePath)
	}

	if strings.Index(dockerCfg.EngineOptions, fmt.Sprintf("SERVERCERT=%s", p.AuthOptions.ServerCertRemotePath)) == -1 {
		t.Fatalf("SERVERCERT option invalid; expected %s", p.AuthOptions.ServerCertRemotePath)
	}
}

func TestMachinePortBoot2Docker(t *testing.T) {
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.Driver{},
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
	p := &Boot2DockerProvisioner{
		Driver: &fakedriver.Driver{},
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
