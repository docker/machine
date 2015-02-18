package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/utils"
)

var (
	validHostnameChars   = `[a-zA-Z0-9\-\.]`
	validHostnamePattern = regexp.MustCompile(`^` + validHostnameChars + `+$`)
)

type (
	Machine struct {
		Name           string `json:"-"`
		DriverName     string
		Driver         drivers.Driver
		CaCertPath     string
		PrivateKeyPath string
		ClientCertPath string
		StorePath      string
	}

	DockerConfig struct {
		EngineConfig     string
		EngineConfigPath string
	}

	machineConfig struct {
		DriverName string
	}
)

func NewMachine(name, driverName, storePath, caCert, privateKey string) (*Machine, error) {
	driver, err := drivers.NewDriver(driverName, name, storePath, caCert, privateKey)
	if err != nil {
		return nil, err
	}

	return &Machine{
		Name:           name,
		DriverName:     driverName,
		Driver:         driver,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
		StorePath:      storePath,
	}, nil
}

func ValidateHostname(name string) (string, error) {
	if !validHostnamePattern.MatchString(name) {
		return name, ErrInvalidHostname
	}
	return name, nil
}

func GenerateClientCertificate(caCertPath, privateKeyPath string) error {
	var (
		org  = "docker-machine"
		bits = 2048
	)

	clientCertPath := filepath.Join(utils.GetMachineDir(), "cert.pem")
	clientKeyPath := filepath.Join(utils.GetMachineDir(), "key.pem")

	if err := os.MkdirAll(utils.GetMachineDir(), 0700); err != nil {
		return err
	}

	log.Debugf("generating client cert: %s", clientCertPath)
	if err := utils.GenerateCert([]string{""}, clientCertPath, clientKeyPath, caCertPath, privateKeyPath, org, bits); err != nil {
		return fmt.Errorf("error generating client cert: %s", err)
	}

	return nil
}

func (m *Machine) ConfigureAuth() error {
	d := m.Driver

	if d.DriverName() == "none" {
		return nil
	}

	ip, err := m.Driver.GetIP()
	if err != nil {
		return err
	}

	serverCertPath := filepath.Join(m.StorePath, "server.pem")
	serverKeyPath := filepath.Join(m.StorePath, "server-key.pem")

	org := m.Name
	bits := 2048

	log.Debugf("generating server cert: %s", serverCertPath)

	if err := utils.GenerateCert([]string{ip}, serverCertPath, serverKeyPath, m.CaCertPath, m.PrivateKeyPath, org, bits); err != nil {
		return fmt.Errorf("error generating server cert: %s", err)
	}

	if err := d.StopDocker(); err != nil {
		return err
	}

	cmd, err := d.GetSSHCommand(fmt.Sprintf("sudo mkdir -p %s", d.GetDockerConfigDir()))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// upload certs and configure TLS auth
	caCert, err := ioutil.ReadFile(m.CaCertPath)
	if err != nil {
		return err
	}

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	machineCaCertPath := path.Join(d.GetDockerConfigDir(), "ca.pem")

	serverCert, err := ioutil.ReadFile(serverCertPath)
	if err != nil {
		return err
	}
	machineServerCertPath := path.Join(d.GetDockerConfigDir(), "server.pem")

	serverKey, err := ioutil.ReadFile(serverKeyPath)
	if err != nil {
		return err
	}
	machineServerKeyPath := path.Join(d.GetDockerConfigDir(), "server-key.pem")

	cmd, err = d.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee -a %s", string(caCert), machineCaCertPath))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee -a %s", string(serverKey), machineServerKeyPath))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee -a %s", string(serverCert), machineServerCertPath))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	dockerUrl, err := m.Driver.GetURL()
	if err != nil {
		return err
	}
	u, err := url.Parse(dockerUrl)
	if err != nil {
		return err
	}
	dockerPort := 2376
	parts := strings.Split(u.Host, ":")
	if len(parts) == 2 {
		dPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		dockerPort = dPort
	}

	cfg := m.generateDockerConfig(dockerPort, machineCaCertPath, machineServerKeyPath, machineServerCertPath)

	cmd, err = d.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee -a %s", cfg.EngineConfig, cfg.EngineConfigPath))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := d.StartDocker(); err != nil {
		return err
	}

	return nil
}

func (m *Machine) generateDockerConfig(dockerPort int, caCertPath string, serverKeyPath string, serverCertPath string) *DockerConfig {
	d := m.Driver
	var (
		daemonOpts    string
		daemonOptsCfg string
		daemonCfg     string
	)

	// TODO @ehazlett: template?
	defaultDaemonOpts := fmt.Sprintf(`--tlsverify \
--tlscacert=%s \
--tlskey=%s \
--tlscert=%s`, caCertPath, serverKeyPath, serverCertPath)

	switch d.DriverName() {

	case "virtualbox", "vmwarefusion", "vmwarevsphere", "hyper-v", "none":
		daemonOpts = fmt.Sprintf("-H tcp://0.0.0.0:%d", dockerPort)
		daemonOptsCfg = path.Join(d.GetDockerConfigDir(), "profile")
		opts := fmt.Sprintf("%s %s", defaultDaemonOpts, daemonOpts)
		daemonCfg = fmt.Sprintf(`EXTRA_ARGS='%s'
CACERT=%s
SERVERCERT=%s
SERVERKEY=%s
DOCKER_TLS=no`, opts, caCertPath, serverKeyPath, serverCertPath)
	default:
		daemonOpts = fmt.Sprintf("--host=unix:///var/run/docker.sock --host=tcp://0.0.0.0:%d", dockerPort)
		daemonOptsCfg = "/etc/default/docker"
		opts := fmt.Sprintf("%s %s", defaultDaemonOpts, daemonOpts)
		daemonCfg = fmt.Sprintf("export DOCKER_OPTS='%s'", opts)
	}

	return &DockerConfig{
		EngineConfig:     daemonCfg,
		EngineConfigPath: daemonOptsCfg,
	}
}

func (m *Machine) Create(name string) error {
	name, err := ValidateHostname(name)
	if err != nil {
		return err
	}

	if err := m.Driver.Create(); err != nil {
		return err
	}

	if err := m.SaveConfig(); err != nil {
		return err
	}

	return nil
}

func (m *Machine) Start() error {
	return m.Driver.Start()
}

func (m *Machine) Stop() error {
	return m.Driver.Stop()
}

func (m *Machine) Upgrade() error {
	return m.Driver.Upgrade()
}

func (m *Machine) Remove(force bool) error {
	if err := m.Driver.Remove(); err != nil {
		if !force {
			return err
		}
	}
	return m.removeStorePath()
}

func (m *Machine) removeStorePath() error {
	file, err := os.Stat(m.StorePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", m.StorePath)
	}
	return os.RemoveAll(m.StorePath)
}

func (m *Machine) GetURL() (string, error) {
	return m.Driver.GetURL()
}

func (m *Machine) LoadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(m.StorePath, "config.json"))
	if err != nil {
		return err
	}

	// First pass: find the driver name and load the driver
	var config machineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	driver, err := drivers.NewDriver(config.DriverName, m.Name, m.StorePath, m.CaCertPath, m.PrivateKeyPath)
	if err != nil {
		return err
	}
	m.Driver = driver

	// Second pass: unmarshal driver config into correct driver
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}

func (m *Machine) SaveConfig() error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(m.StorePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}
