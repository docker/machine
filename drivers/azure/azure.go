package azure

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	azure "github.com/MSOpenTech/azure-sdk-for-go"
	"github.com/MSOpenTech/azure-sdk-for-go/clients/vmClient"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	dockerConfigDir = "/etc/docker"
)

type Driver struct {
	MachineName             string
	SubscriptionID          string
	SubscriptionCert        string
	PublishSettingsFilePath string
	Location                string
	Size                    string
	UserName                string
	UserPassword            string
	Image                   string
	SSHPort                 int
	DockerPort              int
	CaCertPath              string
	PrivateKeyPath          string
	storePath               string
}

func init() {
	drivers.Register("azure", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:  "azure-docker-port",
			Usage: "Azure Docker port",
			Value: 2376,
		},
		cli.StringFlag{
			EnvVar: "AZURE_IMAGE",
			Name:   "azure-image",
			Usage:  "Azure image name. Default is Ubuntu 14.04 LTS x64",
		},
		cli.StringFlag{
			EnvVar: "AZURE_LOCATION",
			Name:   "azure-location",
			Usage:  "Azure location",
			Value:  "West US",
		},
		cli.StringFlag{
			Name:  "azure-password",
			Usage: "Azure user password",
		},
		cli.StringFlag{
			EnvVar: "AZURE_PUBLISH_SETTINGS_FILE",
			Name:   "azure-publish-settings-file",
			Usage:  "Azure publish settings file",
		},
		cli.StringFlag{
			EnvVar: "AZURE_SIZE",
			Name:   "azure-size",
			Usage:  "Azure size",
			Value:  "Small",
		},
		cli.IntFlag{
			Name:  "azure-ssh-port",
			Usage: "Azure SSH port",
			Value: 22,
		},

		cli.StringFlag{
			EnvVar: "AZURE_SUBSCRIPTION_CERT",
			Name:   "azure-subscription-cert",
			Usage:  "Azure subscription cert",
		},
		cli.StringFlag{
			EnvVar: "AZURE_SUBSCRIPTION_ID",
			Name:   "azure-subscription-id",
			Usage:  "Azure subscription ID",
		},
		cli.StringFlag{
			Name:  "azure-username",
			Usage: "Azure username",
			Value: "ubuntu",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	driver := &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}
	return driver, nil
}

func (driver *Driver) DriverName() string {
	return "azure"
}

func (driver *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	driver.SubscriptionID = flags.String("azure-subscription-id")

	cert := flags.String("azure-subscription-cert")
	publishSettings := flags.String("azure-publish-settings-file")
	image := flags.String("azure-image")
	username := flags.String("azure-username")

	if cert != "" {
		if _, err := os.Stat(cert); os.IsNotExist(err) {
			return err
		}
		driver.SubscriptionCert = cert
	}

	if publishSettings != "" {
		if _, err := os.Stat(publishSettings); os.IsNotExist(err) {
			return err
		}
		driver.PublishSettingsFilePath = publishSettings
	}

	if (driver.SubscriptionID == "" || driver.SubscriptionCert == "") && driver.PublishSettingsFilePath == "" {
		return fmt.Errorf("Please specify azure subscription params using options: --azure-subscription-id and --azure-subscription-cert or --azure-publish-settings-file")
	}

	if image == "" {
		driver.Image = "b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB"
	} else {
		driver.Image = image
	}

	driver.Location = flags.String("azure-location")
	driver.Size = flags.String("azure-size")

	if strings.ToLower(username) == "docker" {
		return fmt.Errorf("'docker' is not valid user name for docker host. Please specify another user name")
	}

	driver.UserName = username
	driver.UserPassword = flags.String("azure-password")
	driver.DockerPort = flags.Int("azure-docker-port")
	driver.SSHPort = flags.Int("azure-ssh-port")

	return nil
}

func (driver *Driver) PreCreateCheck() error {
	if err := driver.setUserSubscription(); err != nil {
		return err
	}

	// check azure DNS to make sure name is available
	available, response, err := vmClient.CheckHostedServiceNameAvailability(driver.MachineName)
	if err != nil {
		return err
	}

	if !available {
		return fmt.Errorf(response)
	}

	return nil
}

func (driver *Driver) Create() error {
	if err := driver.setUserSubscription(); err != nil {
		return err
	}

	log.Info("Creating Azure machine...")
	vmConfig, err := vmClient.CreateAzureVMConfiguration(driver.MachineName, driver.Size, driver.Image, driver.Location)
	if err != nil {
		return err
	}

	log.Debug("Generating certificate for Azure...")
	if err := driver.generateCertForAzure(); err != nil {
		return err
	}

	log.Debug("Adding Linux provisioning...")
	vmConfig, err = vmClient.AddAzureLinuxProvisioningConfig(vmConfig, driver.UserName, driver.UserPassword, driver.azureCertPath(), driver.SSHPort)
	if err != nil {
		return err
	}

	log.Debug("Authorizing ports...")
	if err := driver.addDockerEndpoint(vmConfig); err != nil {
		return err
	}

	log.Debug("Creating VM...")
	if err := vmClient.CreateAzureVM(vmConfig, driver.MachineName, driver.Location); err != nil {
		return err
	}

	log.Info("Waiting for SSH...")
	log.Debugf("Host: %s SSH Port: %d", driver.getHostname(), driver.SSHPort)

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", driver.getHostname(), driver.SSHPort)); err != nil {
		return err
	}

	cmd, err := driver.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl -sL https://get.docker.com | sh -; fi")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (driver *Driver) runSSHCommand(command string, retries int) error {
	cmd, err := driver.GetSSHCommand(command)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		if err.Error() == "exit status 255" {
			if retries == 0 {
				return err
			}
			return driver.runSSHCommand(command, retries-1)
		}

		return err
	}

	return nil
}

func (driver *Driver) GetURL() (string, error) {
	url := fmt.Sprintf("tcp://%s:%v", driver.getHostname(), driver.DockerPort)
	return url, nil
}

func (driver *Driver) GetIP() (string, error) {
	return driver.getHostname(), nil
}

func (driver *Driver) GetState() (state.State, error) {
	err := driver.setUserSubscription()
	if err != nil {
		return state.Error, err
	}

	dockerVM, err := vmClient.GetVMDeployment(driver.MachineName, driver.MachineName)
	if err != nil {
		if strings.Contains(err.Error(), "Code: ResourceNotFound") {
			return state.Error, fmt.Errorf("Azure host was not found. Please check your Azure subscription.")
		}

		return state.Error, err
	}

	vmState := dockerVM.RoleInstanceList.RoleInstance[0].PowerState
	switch vmState {
	case "Started":
		return state.Running, nil
	case "Starting":
		return state.Starting, nil
	case "Stopped":
		return state.Stopped, nil
	}

	return state.None, nil
}

func (driver *Driver) Start() error {
	err := driver.setUserSubscription()
	if err != nil {
		return err
	}

	vmState, err := driver.GetState()
	if err != nil {
		return err
	}
	if vmState == state.Running || vmState == state.Starting {
		log.Infof("Host is already running or starting")
		return nil
	}

	log.Debugf("starting %s", driver.MachineName)

	err = vmClient.StartRole(driver.MachineName, driver.MachineName, driver.MachineName)
	if err != nil {
		return err
	}
	err = driver.waitForSSH()
	if err != nil {
		return err
	}
	err = driver.waitForDocker()
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) Stop() error {
	err := driver.setUserSubscription()
	if err != nil {
		return err
	}
	vmState, err := driver.GetState()
	if err != nil {
		return err
	}
	if vmState == state.Stopped {
		log.Infof("Host is already stopped")
		return nil
	}

	log.Debugf("stopping %s", driver.MachineName)

	err = vmClient.ShutdownRole(driver.MachineName, driver.MachineName, driver.MachineName)
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) Remove() error {
	err := driver.setUserSubscription()
	if err != nil {
		return err
	}
	available, _, err := vmClient.CheckHostedServiceNameAvailability(driver.MachineName)
	if err != nil {
		return err
	}
	if available {
		return nil
	}

	log.Debugf("removing %s", driver.MachineName)

	err = vmClient.DeleteHostedService(driver.MachineName)
	if err != nil {
		return err
	}

	return nil
}

func (driver *Driver) Restart() error {
	err := driver.setUserSubscription()
	if err != nil {
		return err
	}
	vmState, err := driver.GetState()
	if err != nil {
		return err
	}
	if vmState == state.Stopped {
		return fmt.Errorf("Host is already stopped, use start command to run it")
	}

	log.Debugf("restarting %s", driver.MachineName)

	err = vmClient.RestartRole(driver.MachineName, driver.MachineName, driver.MachineName)
	if err != nil {
		return err
	}
	err = driver.waitForSSH()
	if err != nil {
		return err
	}
	err = driver.waitForDocker()
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) Kill() error {
	err := driver.setUserSubscription()
	if err != nil {
		return err
	}
	vmState, err := driver.GetState()
	if err != nil {
		return err
	}
	if vmState == state.Stopped {
		log.Infof("Host is already stopped")
		return nil
	}

	log.Debugf("killing %s", driver.MachineName)

	err = vmClient.ShutdownRole(driver.MachineName, driver.MachineName, driver.MachineName)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) StartDocker() error {
	log.Debug("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker start")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) StopDocker() error {
	log.Debug("Stopping Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker stop")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetDockerConfigDir() string {
	return dockerConfigDir
}

func (driver *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	err := driver.setUserSubscription()
	if err != nil {
		return nil, err
	}

	vmState, err := driver.GetState()
	if err != nil {
		return nil, err
	}

	if vmState == state.Stopped {
		return nil, fmt.Errorf("Azure host is stopped. Please start it before using ssh command.")
	}

	return ssh.GetSSHCommand(driver.getHostname(), driver.SSHPort, driver.UserName, driver.sshKeyPath(), args...), nil
}

func (driver *Driver) Upgrade() error {
	log.Debugf("Upgrading Docker")

	cmd, err := driver.GetSSHCommand("sudo apt-get update && sudo apt-get install --upgrade lxc-docker")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
		return err

	}

	return cmd.Run()
}

func generateVMName() string {
	randomID := utils.TruncateID(utils.GenerateRandomID())
	return fmt.Sprintf("docker-host-%s", randomID)
}

func (driver *Driver) setUserSubscription() error {
	if len(driver.PublishSettingsFilePath) != 0 {
		err := azure.ImportPublishSettingsFile(driver.PublishSettingsFilePath)
		if err != nil {
			return err
		}
		return nil
	}
	err := azure.ImportPublishSettings(driver.SubscriptionID, driver.SubscriptionCert)
	if err != nil {
		return err
	}
	return nil
}

func (driver *Driver) addDockerEndpoint(vmConfig *vmClient.Role) error {
	configSets := vmConfig.ConfigurationSets.ConfigurationSet
	if len(configSets) == 0 {
		return fmt.Errorf("no configuration set")
	}
	for i := 0; i < len(configSets); i++ {
		if configSets[i].ConfigurationSetType != "NetworkConfiguration" {
			continue
		}
		ep := vmClient.InputEndpoint{}
		ep.Name = "docker"
		ep.Protocol = "tcp"
		ep.Port = driver.DockerPort
		ep.LocalPort = driver.DockerPort
		configSets[i].InputEndpoints.InputEndpoint = append(configSets[i].InputEndpoints.InputEndpoint, ep)
		log.Debugf("added Docker endpoint (port %d) to configuration", driver.DockerPort)
	}
	return nil
}

func (driver *Driver) waitForSSH() error {
	log.Infof("Waiting for SSH...")
	err := ssh.WaitForTCP(fmt.Sprintf("%s:%v", driver.getHostname(), driver.SSHPort))
	if err != nil {
		return err
	}

	return nil
}

func (driver *Driver) waitForDocker() error {
	log.Infof("Waiting for docker daemon on host to be available...")
	maxRepeats := 48
	url := fmt.Sprintf("%s:%v", driver.getHostname(), driver.DockerPort)
	success := waitForDockerEndpoint(url, maxRepeats)
	if !success {
		return fmt.Errorf("Can not run docker daemon on remote machine. Please try again.")
	}
	return nil
}

func waitForDockerEndpoint(url string, maxRepeats int) bool {
	counter := 0
	for {
		conn, err := net.Dial("tcp", url)
		if err != nil {
			time.Sleep(10 * time.Second)
			counter++
			if counter == maxRepeats {
				return false
			}
			continue
		}
		defer conn.Close()
		break
	}
	return true
}

func (driver *Driver) generateCertForAzure() error {
	if err := ssh.GenerateSSHKey(driver.sshKeyPath()); err != nil {
		return err
	}

	cmd := exec.Command("openssl", "req", "-x509", "-key", driver.sshKeyPath(), "-nodes", "-days", "365", "-newkey", "rsa:2048", "-out", driver.azureCertPath(), "-subj", "/C=AU/ST=Some-State/O=InternetWidgitsPtyLtd/CN=\\*")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (driver *Driver) sshKeyPath() string {
	return filepath.Join(driver.storePath, "id_rsa")
}

func (driver *Driver) publicSSHKeyPath() string {
	return driver.sshKeyPath() + ".pub"
}

func (driver *Driver) azureCertPath() string {
	return filepath.Join(driver.storePath, "azure_cert.pem")
}

func (driver *Driver) getHostname() string {
	return driver.MachineName + ".cloudapp.net"
}
