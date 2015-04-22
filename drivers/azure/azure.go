package azure

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	azure "github.com/MSOpenTech/azure-sdk-for-go"
	"github.com/MSOpenTech/azure-sdk-for-go/clients/vmClient"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	MachineName             string
	SubscriptionID          string
	SubscriptionCert        string
	PublishSettingsFilePath string
	Location                string
	Size                    string
	UserPassword            string
	Image                   string
	SSHUser                 string
	SSHPort                 int
	DockerPort              int
	CaCertPath              string
	PrivateKeyPath          string
	SwarmMaster             bool
	SwarmHost               string
	SwarmDiscovery          string
	storePath               string
}

func init() {
	drivers.Register("azure", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this d adds to
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
	d := &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}
	return d, nil
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "ubuntu"
	}

	return d.SSHUser
}

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Remote
}

func (d *Driver) DriverName() string {
	return "azure"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.SubscriptionID = flags.String("azure-subscription-id")

	cert := flags.String("azure-subscription-cert")
	publishSettings := flags.String("azure-publish-settings-file")
	image := flags.String("azure-image")
	username := flags.String("azure-username")

	if cert != "" {
		if _, err := os.Stat(cert); os.IsNotExist(err) {
			return err
		}
		d.SubscriptionCert = cert
	}

	if publishSettings != "" {
		if _, err := os.Stat(publishSettings); os.IsNotExist(err) {
			return err
		}
		d.PublishSettingsFilePath = publishSettings
	}

	if (d.SubscriptionID == "" || d.SubscriptionCert == "") && d.PublishSettingsFilePath == "" {
		return errors.New("Please specify azure subscription params using options: --azure-subscription-id and --azure-subscription-cert or --azure-publish-settings-file")
	}

	if image == "" {
		d.Image = "b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB"
	} else {
		d.Image = image
	}

	d.Location = flags.String("azure-location")
	d.Size = flags.String("azure-size")

	if strings.ToLower(username) == "docker" {
		return errors.New("'docker' is not valid user name for docker host. Please specify another user name")
	}

	d.SSHUser = username
	d.UserPassword = flags.String("azure-password")
	d.DockerPort = flags.Int("azure-docker-port")
	d.SSHPort = flags.Int("azure-ssh-port")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	// check azure DNS to make sure name is available
	available, response, err := vmClient.CheckHostedServiceNameAvailability(d.MachineName)
	if err != nil {
		return err
	}

	if !available {
		return errors.New(response)
	}

	return nil
}

func (d *Driver) Create() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	log.Info("Creating Azure machine...")
	vmConfig, err := vmClient.CreateAzureVMConfiguration(d.MachineName, d.Size, d.Image, d.Location)
	if err != nil {
		return err
	}

	log.Debug("Generating certificate for Azure...")
	if err := d.generateCertForAzure(); err != nil {
		return err
	}

	log.Debug("Adding Linux provisioning...")
	vmConfig, err = vmClient.AddAzureLinuxProvisioningConfig(vmConfig, d.GetSSHUsername(), d.UserPassword, d.azureCertPath(), d.SSHPort)
	if err != nil {
		return err
	}

	log.Debug("Authorizing ports...")
	if err := d.addDockerEndpoint(vmConfig); err != nil {
		return err
	}

	log.Debug("Creating VM...")
	if err := vmClient.CreateAzureVM(vmConfig, d.MachineName, d.Location); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	url := fmt.Sprintf("tcp://%s:%v", d.getHostname(), d.DockerPort)
	return url, nil
}

func (d *Driver) GetIP() (string, error) {
	return d.getHostname(), nil
}

func (d *Driver) GetState() (state.State, error) {
	if err := d.setUserSubscription(); err != nil {
		return state.Error, err
	}

	dockerVM, err := vmClient.GetVMDeployment(d.MachineName, d.MachineName)
	if err != nil {
		if strings.Contains(err.Error(), "Code: ResourceNotFound") {
			return state.Error, errors.New("Azure host was not found. Please check your Azure subscription.")
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

func (d *Driver) Start() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Running || vmState == state.Starting {
		log.Infof("Host is already running or starting")
		return nil
	}

	log.Debugf("starting %s", d.MachineName)

	if err := vmClient.StartRole(d.MachineName, d.MachineName, d.MachineName); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Stopped {
		log.Infof("Host is already stopped")
		return nil
	}

	log.Debugf("stopping %s", d.MachineName)

	return vmClient.ShutdownRole(d.MachineName, d.MachineName, d.MachineName)
}

func (d *Driver) Remove() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}
	if available, _, err := vmClient.CheckHostedServiceNameAvailability(d.MachineName); err != nil {
		return err
	} else if available {
		return nil
	}

	log.Debugf("removing %s", d.MachineName)

	return vmClient.DeleteHostedService(d.MachineName)
}

func (d *Driver) Restart() error {
	err := d.setUserSubscription()
	if err != nil {
		return err
	}
	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Stopped {
		return errors.New("Host is already stopped, use start command to run it")
	}

	log.Debugf("restarting %s", d.MachineName)

	if err := vmClient.RestartRole(d.MachineName, d.MachineName, d.MachineName); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Stopped {
		log.Infof("Host is already stopped")
		return nil
	}

	log.Debugf("killing %s", d.MachineName)

	return vmClient.ShutdownRole(d.MachineName, d.MachineName, d.MachineName)
}

func generateVMName() string {
	randomID := utils.TruncateID(utils.GenerateRandomID())
	return fmt.Sprintf("docker-host-%s", randomID)
}

func (d *Driver) setUserSubscription() error {
	if d.PublishSettingsFilePath != "" {
		return azure.ImportPublishSettingsFile(d.PublishSettingsFilePath)
	}
	return azure.ImportPublishSettings(d.SubscriptionID, d.SubscriptionCert)
}

func (d *Driver) addDockerEndpoint(vmConfig *vmClient.Role) error {
	configSets := vmConfig.ConfigurationSets.ConfigurationSet
	if len(configSets) == 0 {
		return errors.New("no configuration set")
	}
	for i := 0; i < len(configSets); i++ {
		if configSets[i].ConfigurationSetType != "NetworkConfiguration" {
			continue
		}
		ep := vmClient.InputEndpoint{
			Name:      "docker",
			Protocol:  "tcp",
			Port:      d.DockerPort,
			LocalPort: d.DockerPort}
		configSets[i].InputEndpoints.InputEndpoint = append(configSets[i].InputEndpoints.InputEndpoint, ep)
		log.Debugf("added Docker endpoint (port %d) to configuration", d.DockerPort)
	}
	return nil
}

func (d *Driver) generateCertForAzure() error {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	cmd := exec.Command("openssl", "req", "-x509", "-key", d.sshKeyPath(), "-nodes", "-days", "365", "-newkey", "rsa:2048", "-out", d.azureCertPath(), "-subj", "/C=AU/ST=Some-State/O=InternetWidgitsPtyLtd/CN=\\*")
	return cmd.Run()
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) azureCertPath() string {
	return filepath.Join(d.storePath, "azure_cert.pem")
}

func (d *Driver) getHostname() string {
	return d.MachineName + ".cloudapp.net"
}
