package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/utils"
)

var (
	validHostNameChars   = `[a-zA-Z0-9_]`
	validHostNamePattern = regexp.MustCompile(`^` + validHostNameChars + `+$`)
)

var HostOSTypeNotKnown = errors.New("Host OS type is not known")

type Host struct {
	Name           			string `json:"-"`
	DriverName     			string
	Driver         			drivers.Driver
	CaCertPath     			string
	ServerCertPath 			string
	ServerKeyPath  			string
	PrivateKeyPath 			string
	ClientCertPath 			string
	storePath      			string
	RemoteDockerOpts 		string
	HostOSType				string
	HostDockerConfigFile	string
	HostDockerRestartCmd	string
	HostDockerConfigKey		string
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

func NewHost(name, driverName, storePath, remoteDockerOpts, caCert, privateKey string) (*Host, error) {
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
		storePath:      storePath,
		RemoteDockerOpts: 	remoteDockerOpts,
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
		return name, fmt.Errorf("Invalid host name %q, it must match %s", name, validHostNamePattern)
	}
	return name, nil
}

func (h *Host) GenerateCertificates() error {
	var (
		caPathExists     bool
		privateKeyExists bool
		org              = "docker-machine"
		bits             = 2048
	)

	ip, err := h.Driver.GetIP()
	if err != nil {
		return err
	}

	caCertPath := filepath.Join(h.storePath, "ca.pem")
	privateKeyPath := filepath.Join(h.storePath, "private.pem")

	if _, err := os.Stat(h.CaCertPath); os.IsNotExist(err) {
		caPathExists = false
	} else {
		caPathExists = true
	}

	if _, err := os.Stat(h.PrivateKeyPath); os.IsNotExist(err) {
		privateKeyExists = false
	} else {
		privateKeyExists = true
	}

	if !caPathExists && !privateKeyExists {
		log.Debugf("generating self-signed CA cert: %s", caCertPath)
		if err := utils.GenerateCACert(caCertPath, privateKeyPath, org, bits); err != nil {
			return fmt.Errorf("error generating self-signed CA cert: %s", err)
		}
	} else {
		if err := utils.CopyFile(h.CaCertPath, caCertPath); err != nil {
			return fmt.Errorf("unable to copy CA cert: %s", err)
		}
		if err := utils.CopyFile(h.PrivateKeyPath, privateKeyPath); err != nil {
			return fmt.Errorf("unable to copy private key: %s", err)
		}
	}

	serverCertPath := filepath.Join(h.storePath, "server.pem")
	serverKeyPath := filepath.Join(h.storePath, "server-key.pem")

	log.Debugf("generating server cert: %s", serverCertPath)

	if err := utils.GenerateCert([]string{ip}, serverCertPath, serverKeyPath, caCertPath, privateKeyPath, org, bits); err != nil {
		return fmt.Errorf("error generating server cert: %s", err)
	}

	clientCertPath := filepath.Join(h.storePath, "cert.pem")
	clientKeyPath := filepath.Join(h.storePath, "key.pem")
	log.Debugf("generating client cert: %s", clientCertPath)
	if err := utils.GenerateCert([]string{""}, clientCertPath, clientKeyPath, caCertPath, privateKeyPath, org, bits); err != nil {
		return fmt.Errorf("error generating client cert: %s", err)
	}

	return nil
}

func (h *Host) ConfigureAuth() error {
	d := h.Driver

	if d.DriverName() == "none" {
		return nil
	}

	log.Debugf("generating certificates for %s", h.Name)
	if err := h.GenerateCertificates(); err != nil {
		return err
	}

	serverCertPath := filepath.Join(h.storePath, "server.pem")
	caCertPath := filepath.Join(h.storePath, "ca.pem")
	serverKeyPath := filepath.Join(h.storePath, "server-key.pem")

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
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return err
	}

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	machineCaCertPath := fmt.Sprintf("%s/ca.pem", d.GetDockerConfigDir())

	serverCert, err := ioutil.ReadFile(serverCertPath)
	if err != nil {
		return err
	}
	machineServerCertPath := fmt.Sprintf("%s/server.pem", d.GetDockerConfigDir())

	serverKey, err := ioutil.ReadFile(serverKeyPath)
	if err != nil {
		return err
	}
	machineServerKeyPath := fmt.Sprintf("%s/server-key.pem", d.GetDockerConfigDir())

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

	var (
		daemonOpts    string
		daemonOptsCfg string
		daemonCfg     string
	)

	// TODO @ehazlett: template?
	defaultDaemonOpts := fmt.Sprintf(`--tlsverify \
--tlscacert=%s \
--tlskey=%s \
--tlscert=%s`, machineCaCertPath, machineServerKeyPath, machineServerCertPath)

	switch d.DriverName() {
	case "virtualbox", "vmwarefusion", "vmwarevsphere":
		daemonOpts = "-H tcp://0.0.0.0:2376"
		daemonOptsCfg = filepath.Join(d.GetDockerConfigDir(), "profile")
		opts := fmt.Sprintf("%s %s", defaultDaemonOpts, daemonOpts)
		daemonCfg = fmt.Sprintf(`EXTRA_ARGS='%s'
CACERT=%s
SERVERCERT=%s
SERVERKEY=%s
DOCKER_TLS=no`, opts, machineCaCertPath, machineServerCertPath, machineServerKeyPath)
	default:
		daemonOpts = "--host=unix:///var/run/docker.sock --host=tcp://0.0.0.0:2376"
		daemonOptsCfg = "/etc/default/docker"
		opts := fmt.Sprintf("%s %s", defaultDaemonOpts, daemonOpts)
		daemonCfg = fmt.Sprintf("export DOCKER_OPTS='%s'", opts)
	}
	cmd, err = d.GetSSHCommand(fmt.Sprintf("echo \"%s\" | sudo tee -a %s", daemonCfg, daemonOptsCfg))
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

func (h *Host) Create(name string) error {
	if err := h.Driver.Create(); err != nil {
		return err
	}

	if err := h.setHostOSDockerAttributes(); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	if err := h.ConfigureRemoteDockerDaemon(); err != nil {
		return err
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

func(h *Host) setHostOSDockerAttributes() error {

	hostOsType, err := h.getHostOSType()
	if err != nil {
		return err
	} else {
		h.HostOSType = hostOsType
	}

	hostDockerConfigFile, err := h.getDockerConfigFilePath()
	if err != nil {
		return err
	} else {
		h.HostDockerConfigFile = hostDockerConfigFile
	}

	
	hostDockerRestartCmd, err := h.getDockerRestartCmd()
	if err != nil {
		return err
	} else {
		h.HostDockerRestartCmd = hostDockerRestartCmd
	}

	
	hostDockerConfigKey, err := h.getDockerOptsVarName()
	if err != nil {
		return err
	} else {
		h.HostDockerConfigKey = hostDockerConfigKey
	}


	return nil

}

func(h *Host) getHostOSType() (string, error) {
	// Will have to do somme h.Driver.GetSSHComand to determine the OS type
	return "boot2docker", nil
}

func(h *Host) getDockerConfigFilePath() (string, error) {
	
	switch h.HostOSType  {
    case "boot2docker":
    	return "/var/lib/boot2docker/profile", nil
    default:
    	return "", HostOSTypeNotKnown
    }
}

func(h *Host) getDockerRestartCmd() (string, error) {

	switch h.HostOSType  {
    case "boot2docker":
    	return "sudo /etc/init.d/docker restart", nil
    default:
    	return "", HostOSTypeNotKnown
    }
}

func(h *Host) getDockerOptsVarName() (string, error) {

	switch h.HostOSType  {
    case "boot2docker":
    	return "EXTRA_ARGS", nil
    default:
    	return "", HostOSTypeNotKnown
    }
}

func(h *Host) WriteDaemonKVPConfig(configKey, configValue string) error {

	kvpString := configKey + "=" + configValue
	sshCmd := fmt.Sprintf("echo \"%s\" | sudo tee -a %s", kvpString, h.HostDockerConfigFile)
	cmd, err := h.Driver.GetSSHCommand(sshCmd)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func(h *Host) RestartHostDockerDaemon() error {

	cmd, err := h.Driver.GetSSHCommand(h.HostDockerRestartCmd)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (h *Host) ConfigureRemoteDockerDaemon() error {

	err := h.WriteDaemonKVPConfig(h.HostDockerConfigKey, h.RemoteDockerOpts)
	if err != nil {
		return err
	}

	err = h.RestartHostDockerDaemon()
	if err != nil {
		return err
	}

	// Cleaning existing values assuming that the remote
	// h.Driver.

	// Writing the new value of DOCKER_OPTS at the end of the file

	// Restart Docker

	return nil
}
