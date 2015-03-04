package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/utils"
)

var (
	validHostNameChars   = `[a-zA-Z0-9\-\.]`
	validHostNamePattern = regexp.MustCompile(`^` + validHostNameChars + `+$`)
	ErrInvalidHostname   = errors.New("Invalid hostname specified")
)

const (
	swarmDockerImage              = "swarm:latest"
	swarmDiscoveryServiceEndpoint = "https://discovery-stage.hub.docker.com/v1"
)

type Host struct {
	Name           string `json:"-"`
	DriverName     string
	Driver         drivers.Driver
	CaCertPath     string
	ServerCertPath string
	ServerKeyPath  string
	PrivateKeyPath string
	ClientCertPath string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
	storePath      string
}

type DockerConfig struct {
	EngineConfig     string
	EngineConfigPath string
}

type hostConfig struct {
	DriverName string
}

func waitForDocker(addr string) error {
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second*5)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		conn.Close()
		break
	}
	return nil
}

func NewHost(name, driverName, storePath, caCert, privateKey string, swarmMaster bool, swarmHost string, swarmDiscovery string) (*Host, error) {
	driver, err := drivers.NewDriver(driverName, name, storePath, caCert, privateKey)
	if err != nil {
		return nil, err
	}
	return &Host{
		Name:           name,
		DriverName:     driverName,
		Driver:         driver,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
		SwarmMaster:    swarmMaster,
		SwarmHost:      swarmHost,
		SwarmDiscovery: swarmDiscovery,
		storePath:      storePath,
	}, nil
}

func LoadHost(name string, storePath string) (*Host, error) {
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Host %q does not exist", name)
	}

	host := &Host{Name: name, storePath: storePath}
	if err := host.LoadConfig(); err != nil {
		return nil, err
	}
	return host, nil
}

func ValidateHostName(name string) (string, error) {
	if !validHostNamePattern.MatchString(name) {
		return name, ErrInvalidHostname
	}
	return name, nil
}

func (h *Host) ConfigureSwarm(discovery string, master bool, host string, addr string) error {
	d := h.Driver

	if d.DriverName() == "none" {
		return nil
	}

	if addr == "" {
		ip, err := d.GetIP()
		if err != nil {
			return err
		}
		// TODO: remove hardcoded port
		addr = fmt.Sprintf("%s:2376", ip)
	}

	basePath := d.GetDockerConfigDir()
	tlsCaCert := path.Join(basePath, "ca.pem")
	tlsCert := path.Join(basePath, "server.pem")
	tlsKey := path.Join(basePath, "server-key.pem")
	masterArgs := fmt.Sprintf("--tlsverify --tlscacert=%s --tlscert=%s --tlskey=%s -H %s --strategy random %s",
		tlsCaCert, tlsCert, tlsKey, host, discovery)
	nodeArgs := fmt.Sprintf("--addr %s %s", addr, discovery)

	u, err := url.Parse(host)
	if err != nil {
		return err
	}

	parts := strings.Split(u.Host, ":")
	port := parts[1]

	if err := waitForDocker(addr); err != nil {
		return err
	}

	cmd, err := d.GetSSHCommand(fmt.Sprintf("sudo docker pull %s", swarmDockerImage))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// if master start master agent
	if master {
		log.Debug("launching swarm master")
		log.Debugf("master args: %s", masterArgs)
		cmd, err = d.GetSSHCommand(fmt.Sprintf("sudo docker run -d -p %s:%s --restart=always --name swarm-agent-master -v %s:%s %s manage %s",
			port, port, d.GetDockerConfigDir(), d.GetDockerConfigDir(), swarmDockerImage, masterArgs))
		if err != nil {
			return err
		}
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// start node agent
	log.Debug("launching swarm node")
	log.Debugf("node args: %s", nodeArgs)
	cmd, err = d.GetSSHCommand(fmt.Sprintf("sudo docker run -d --restart=always --name swarm-agent -v %s:%s %s join %s",
		d.GetDockerConfigDir(), d.GetDockerConfigDir(), swarmDockerImage, nodeArgs))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (h *Host) ConfigureAuth() error {
	d := h.Driver

	if d.DriverName() == "none" {
		return nil
	}

	// copy certs to client dir for docker client
	machineDir := filepath.Join(utils.GetMachineDir(), h.Name)
	if err := utils.CopyFile(h.CaCertPath, filepath.Join(machineDir, "ca.pem")); err != nil {
		log.Fatalf("Error copying ca.pem to machine dir: %s", err)
	}

	clientCertPath := filepath.Join(utils.GetMachineCertDir(), "cert.pem")
	if err := utils.CopyFile(clientCertPath, filepath.Join(machineDir, "cert.pem")); err != nil {
		log.Fatalf("Error copying cert.pem to machine dir: %s", err)
	}

	clientKeyPath := filepath.Join(utils.GetMachineCertDir(), "key.pem")
	if err := utils.CopyFile(clientKeyPath, filepath.Join(machineDir, "key.pem")); err != nil {
		log.Fatalf("Error copying key.pem to machine dir: %s", err)
	}

	var (
		ip         = ""
		ipErr      error
		maxRetries = 4
	)

	for i := 0; i < maxRetries; i++ {
		ip, ipErr = h.Driver.GetIP()
		if ip != "" {
			break
		}
		log.Debugf("waiting for ip: %s", ipErr)
		time.Sleep(5 * time.Second)
	}

	if ipErr != nil {
		return ipErr
	}

	if ip == "" {
		return fmt.Errorf("unable to get machine IP")
	}

	serverCertPath := filepath.Join(h.storePath, "server.pem")
	serverKeyPath := filepath.Join(h.storePath, "server-key.pem")

	org := h.Name
	bits := 2048

	log.Debugf("generating server cert: %s ca-key=%s private-key=%s org=%s",
		serverCertPath,
		h.CaCertPath,
		h.PrivateKeyPath,
		org,
	)

	if err := utils.GenerateCert([]string{ip}, serverCertPath, serverKeyPath, h.CaCertPath, h.PrivateKeyPath, org, bits); err != nil {
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
	caCert, err := ioutil.ReadFile(h.CaCertPath)
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

	dockerUrl, err := h.Driver.GetURL()
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

	cfg := h.generateDockerConfig(dockerPort, machineCaCertPath, machineServerKeyPath, machineServerCertPath)

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

func (h *Host) generateDockerConfig(dockerPort int, caCertPath string, serverKeyPath string, serverCertPath string) *DockerConfig {
	d := h.Driver
	var (
		daemonOpts    string
		daemonOptsCfg string
		daemonCfg     string
		swarmLabels   = []string{}
	)

	swarmLabels = append(swarmLabels, fmt.Sprintf("--label=provider=%s", h.Driver.DriverName()))

	defaultDaemonOpts := fmt.Sprintf(`--tlsverify --tlscacert=%s --tlskey=%s --tlscert=%s %s`,
		caCertPath,
		serverKeyPath,
		serverCertPath,
		strings.Join(swarmLabels, " "),
	)

	switch d.DriverName() {
	case "virtualbox", "vmwarefusion", "vmwarevsphere", "hyper-v":
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
		daemonCfg = fmt.Sprintf("export DOCKER_OPTS=\\\"%s\\\"", opts)
	}

	return &DockerConfig{
		EngineConfig:     daemonCfg,
		EngineConfigPath: daemonOptsCfg,
	}
}

func (h *Host) Create(name string) error {
	name, err := ValidateHostName(name)
	if err != nil {
		return err
	}

	// create the instance
	if err := h.Driver.Create(); err != nil {
		return err
	}

	// install docker
	if err := h.Provision(); err != nil {
		return err
	}

	// save to store
	if err := h.SaveConfig(); err != nil {
		return err
	}

	return nil
}

func (h *Host) Provision() error {
	// "local" providers use b2d; no provisioning necessary
	switch h.Driver.DriverName() {
	case "none", "virtualbox", "vmwarefusion", "vmwarevsphere":
		return nil
	}

	// install docker - until cloudinit we use ubuntu everywhere so we
	// just install it using the docker repos
	cmd, err := h.Driver.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl -sSL https://get.docker.com | sh -; fi")
	if err != nil {
		return err
	}

	// HACK: the script above will output debug to stderr; we save it and
	// then check if the command returned an error; if so, we show the debug

	var buf bytes.Buffer
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error installing docker: %s\n%s\n", err, string(buf.Bytes()))
	}

	return nil
}

func (h *Host) Start() error {
	return h.Driver.Start()
}

func (h *Host) Stop() error {
	return h.Driver.Stop()
}

func (h *Host) Upgrade() error {
	return h.Driver.Upgrade()
}

func (h *Host) Remove(force bool) error {
	if err := h.Driver.Remove(); err != nil {
		if !force {
			return err
		}
	}
	return h.removeStorePath()
}

func (h *Host) removeStorePath() error {
	file, err := os.Stat(h.storePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", h.storePath)
	}
	return os.RemoveAll(h.storePath)
}

func (h *Host) GetURL() (string, error) {
	return h.Driver.GetURL()
}

func (h *Host) LoadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(h.storePath, "config.json"))
	if err != nil {
		return err
	}

	// First pass: find the driver name and load the driver
	var config hostConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	driver, err := drivers.NewDriver(config.DriverName, h.Name, h.storePath, h.CaCertPath, h.PrivateKeyPath)
	if err != nil {
		return err
	}
	h.Driver = driver

	// Second pass: unmarshal driver config into correct driver
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	return nil
}

func (h *Host) SaveConfig() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(h.storePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}
