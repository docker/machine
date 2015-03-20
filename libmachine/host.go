package libmachine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

var (
	validHostNameChars   = `[a-zA-Z0-9\-\.]`
	validHostNamePattern = regexp.MustCompile(`^` + validHostNameChars + `+$`)
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
	PrivateKeyPath string
	ServerCertPath string
	ServerKeyPath  string
	ClientCertPath string
	StorePath      string
	EngineOptions  *engine.EngineOptions
	SwarmOptions   *swarm.SwarmOptions
	// deprecated options; these are left to assist in config migrations
	SwarmHost      string
	SwarmMaster    bool
	SwarmDiscovery string
}

type HostOptions struct {
	Driver        string
	Memory        int
	Disk          int
	DriverOptions drivers.DriverOptions
	EngineOptions *engine.EngineOptions
	SwarmOptions  *swarm.SwarmOptions
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

func NewHost(name, driverName, StorePath, caCert, privateKey string, engineOptions *engine.EngineOptions, swarmOptions *swarm.SwarmOptions) (*Host, error) {
	driver, err := drivers.NewDriver(driverName, name, StorePath, caCert, privateKey)
	if err != nil {
		return nil, err
	}
	return &Host{
		Name:           name,
		DriverName:     driverName,
		Driver:         driver,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
		EngineOptions:  engineOptions,
		SwarmOptions:   swarmOptions,
		StorePath:      StorePath,
	}, nil
}

func LoadHost(name string, StorePath string) (*Host, error) {
	if _, err := os.Stat(StorePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Host %q does not exist", name)
	}

	host := &Host{Name: name, StorePath: StorePath}
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

func (h *Host) GetDockerConfigDir() (string, error) {
	// TODO: this will be refactored in https://github.com/docker/machine/issues/699
	switch h.Driver.GetProviderType() {
	case provider.Local:
		return "/var/lib/boot2docker", nil
	case provider.Remote:
		return "/etc/docker", nil
	case provider.None:
		return "", nil
	default:
		return "", ErrUnknownProviderType
	}
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

	basePath, err := h.GetDockerConfigDir()
	if err != nil {
		return err
	}

	tlsCaCert := path.Join(basePath, "ca.pem")
	tlsCert := path.Join(basePath, "server.pem")
	tlsKey := path.Join(basePath, "server-key.pem")
	masterArgs := fmt.Sprintf("--tlsverify --tlscacert=%s --tlscert=%s --tlskey=%s -H %s %s",
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

	cmd, err := h.GetSSHCommand(fmt.Sprintf("sudo docker pull %s", swarmDockerImage))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	dockerDir, err := h.GetDockerConfigDir()
	if err != nil {
		return err
	}

	// if master start master agent
	if master {
		log.Debug("launching swarm master")
		log.Debugf("master args: %s", masterArgs)
		cmd, err = h.GetSSHCommand(fmt.Sprintf("sudo docker run -d -p %s:%s --restart=always --name swarm-agent-master -v %s:%s %s manage %s",
			port, port, dockerDir, dockerDir, swarmDockerImage, masterArgs))
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
	cmd, err = h.GetSSHCommand(fmt.Sprintf("sudo docker run -d --restart=always --name swarm-agent -v %s:%s %s join %s",
		dockerDir, dockerDir, swarmDockerImage, nodeArgs))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (h *Host) StartDocker() error {
	log.Debug("Starting Docker...")

	var (
		cmd *exec.Cmd
		err error
	)

	switch h.Driver.GetProviderType() {
	case provider.Local:
		cmd, err = h.GetSSHCommand("sudo /etc/init.d/docker start")
	case provider.Remote:
		cmd, err = h.GetSSHCommand("sudo service docker start")
	default:
		return ErrUnknownProviderType
	}

	if err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (h *Host) StopDocker() error {
	log.Debug("Stopping Docker...")

	var (
		cmd *exec.Cmd
		err error
	)

	switch h.Driver.GetProviderType() {
	case provider.Local:
		cmd, err = h.GetSSHCommand("if [ -e /var/run/docker.pid  ] && [ -d /proc/$(cat /var/run/docker.pid)  ]; then sudo /etc/init.d/docker stop ; exit 0; fi")
	case provider.Remote:
		cmd, err = h.GetSSHCommand("sudo service docker stop")
	default:
		return ErrUnknownProviderType
	}

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

	serverCertPath := filepath.Join(h.StorePath, "server.pem")
	serverKeyPath := filepath.Join(h.StorePath, "server-key.pem")

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

	if err := h.StopDocker(); err != nil {
		return err
	}

	dockerDir, err := h.GetDockerConfigDir()
	if err != nil {
		return err
	}

	cmd, err := h.GetSSHCommand(fmt.Sprintf("sudo mkdir -p %s", dockerDir))
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
	machineCaCertPath := path.Join(dockerDir, "ca.pem")

	serverCert, err := ioutil.ReadFile(serverCertPath)
	if err != nil {
		return err
	}
	machineServerCertPath := path.Join(dockerDir, "server.pem")

	serverKey, err := ioutil.ReadFile(serverKeyPath)
	if err != nil {
		return err
	}
	machineServerKeyPath := path.Join(dockerDir, "server-key.pem")

	cmd, err = h.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee %s", string(caCert), machineCaCertPath))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = h.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee %s", string(serverKey), machineServerKeyPath))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = h.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee %s", string(serverCert), machineServerCertPath))
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

	cfg, err := h.generateDockerConfig(dockerPort, machineCaCertPath, machineServerKeyPath, machineServerCertPath)
	if err != nil {
		return err
	}

	cmd, err = h.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee -a %s", cfg.EngineConfig, cfg.EngineConfigPath))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := h.StartDocker(); err != nil {
		return err
	}

	return nil
}

func (h *Host) generateDockerConfig(dockerPort int, caCertPath string, serverKeyPath string, serverCertPath string) (*DockerConfig, error) {
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

	dockerDir, err := h.GetDockerConfigDir()
	if err != nil {
		return nil, err
	}

	switch d.DriverName() {
	case "virtualbox", "vmwarefusion", "vmwarevsphere", "hyper-v":
		daemonOpts = fmt.Sprintf("-H tcp://0.0.0.0:%d", dockerPort)
		daemonOptsCfg = path.Join(dockerDir, "profile")
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
	}, nil
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

	// save to store
	if err := h.SaveConfig(); err != nil {
		return err
	}

	// set hostname
	if err := h.SetHostname(); err != nil {
		return err
	}

	// install docker
	if err := h.Provision(); err != nil {
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

	if err := WaitForSSH(h); err != nil {
		return err
	}

	// install docker - until cloudinit we use ubuntu everywhere so we
	// just install it using the docker repos
	cmd, err := h.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl -sSL https://get.docker.com | sh -; fi")
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

func (h *Host) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	addr, err := h.Driver.GetSSHHostname()
	if err != nil {
		return nil, err
	}

	user := h.Driver.GetSSHUsername()

	port, err := h.Driver.GetSSHPort()
	if err != nil {
		return nil, err
	}

	keyPath := h.Driver.GetSSHKeyPath()

	cmd := ssh.GetSSHCommand(addr, port, user, keyPath, args...)
	return cmd, nil
}

func (h *Host) SetHostname() error {
	var (
		cmd *exec.Cmd
		err error
	)

	log.Debugf("setting hostname for provider type %s: %s",
		h.Driver.GetProviderType(),
		h.Name,
	)

	switch h.Driver.GetProviderType() {
	case provider.None:
		return nil
	case provider.Local:
		cmd, err = h.GetSSHCommand(fmt.Sprintf(
			"sudo hostname %s && echo \"%s\" | sudo tee /var/lib/boot2docker/etc/hostname",
			h.Name,
			h.Name,
		))
	case provider.Remote:
		cmd, err = h.GetSSHCommand(fmt.Sprintf(
			"echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts && sudo hostname %s && echo \"%s\" | sudo tee /etc/hostname",
			h.Name,
			h.Name,
			h.Name,
		))
	default:
		return ErrUnknownProviderType
	}

	if err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (h *Host) MachineInState(desiredState state.State) func() bool {
	return func() bool {
		currentState, err := h.Driver.GetState()
		if err != nil {
			log.Debugf("Error getting machine state: %s", err)
		}
		if currentState == desiredState {
			return true
		}
		return false
	}
}

func (h *Host) Start() error {
	if err := h.Driver.Start(); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return utils.WaitFor(h.MachineInState(state.Running))
}

func (h *Host) Stop() error {
	if err := h.Driver.Stop(); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return utils.WaitFor(h.MachineInState(state.Stopped))
}

func (h *Host) Kill() error {
	if err := h.Driver.Stop(); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return utils.WaitFor(h.MachineInState(state.Stopped))
}

func (h *Host) Restart() error {
	if h.MachineInState(state.Running)() {
		if err := h.Stop(); err != nil {
			return err
		}

		if err := utils.WaitFor(h.MachineInState(state.Stopped)); err != nil {
			return err
		}
	}

	if err := h.Start(); err != nil {
		return err
	}

	if err := utils.WaitFor(h.MachineInState(state.Running)); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return nil
}

func (h *Host) Upgrade() error {
	// TODO: refactor to provisioner
	return fmt.Errorf("centralized upgrade coming in the provisioner")
}

func (h *Host) Remove(force bool) error {
	if err := h.Driver.Remove(); err != nil {
		if !force {
			return err
		}
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return h.removeStorePath()
}

func (h *Host) removeStorePath() error {
	file, err := os.Stat(h.StorePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", h.StorePath)
	}
	return os.RemoveAll(h.StorePath)
}

func (h *Host) GetURL() (string, error) {
	return h.Driver.GetURL()
}

func (h *Host) LoadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(h.StorePath, "config.json"))
	if err != nil {
		return err
	}

	// First pass: find the driver name and load the driver
	var config hostConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	driver, err := drivers.NewDriver(config.DriverName, h.Name, h.StorePath, h.CaCertPath, h.PrivateKeyPath)
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

	if err := ioutil.WriteFile(filepath.Join(h.StorePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}

func sshAvailableFunc(h *Host) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForSSH function...")
		hostname, err := h.Driver.GetSSHHostname()
		if err != nil {
			log.Debugf("Error getting IP address waiting for SSH: %s", err)
			return false
		}
		port, err := h.Driver.GetSSHPort()
		if err != nil {
			log.Debugf("Error getting SSH port: %s", err)
			return false
		}
		if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", hostname, port)); err != nil {
			log.Debugf("Error waiting for TCP waiting for SSH: %s", err)
			return false
		}
		cmd, err := h.GetSSHCommand("exit 0")
		if err != nil {
			log.Debugf("Error getting ssh command 'exit 0' : %s", err)
			return false
		}
		if err := cmd.Run(); err != nil {
			log.Debugf("Error running ssh command 'exit 0' : %s", err)
			return false
		}
		return true
	}
}

func WaitForSSH(h *Host) error {
	if err := utils.WaitFor(sshAvailableFunc(h)); err != nil {
		return fmt.Errorf("Too many retries.  Last error: %s", err)
	}
	return nil
}
