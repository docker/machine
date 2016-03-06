package azure

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/hostedservice"
	"github.com/Azure/azure-sdk-for-go/management/virtualmachine"
	"github.com/Azure/azure-sdk-for-go/management/vmutils"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	SubscriptionID          string
	SubscriptionCert        string
	PublishSettingsFilePath string
	Location                string
	Size                    string
	UserPassword            string
	Image                   string
	OS                      string
	DockerPort              int
	DockerSwarmMasterPort   int
}

const (
	defaultWindowsImage    = "cx-win-2016-2"
	defaultLinuxImage      = "b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-15_10-amd64-server-20151116.1-en-us-30GB"
	defaultDockerPort      = 2376
	defaultSwarmMasterPort = 3376
	defaultLocation        = "West US"
	defaultSize            = "Small"
	defaultSSHPort         = 22
	defaultSSHUsername     = "ubuntu"
)

// GetCreateFlags registers the flags this d adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.IntFlag{
			Name:  "azure-docker-port",
			Usage: "Azure Docker port",
			Value: defaultDockerPort,
		},
		mcnflag.IntFlag{
			Name:  "azure-docker-swarm-master-port",
			Usage: "Azure Docker Swarm master port",
			Value: defaultSwarmMasterPort,
		},
		mcnflag.StringFlag{
			EnvVar: "AZURE_IMAGE",
			Name:   "azure-image",
			Usage:  "Azure image name. Default is Ubuntu 14.04 LTS x64",
		},
		mcnflag.StringFlag{
			EnvVar: "AZURE_LOCATION",
			Name:   "azure-location",
			Usage:  "Azure location",
			Value:  defaultLocation,
		},
		mcnflag.StringFlag{
			Name:  "azure-password",
			Usage: "Azure user password",
		},
		mcnflag.StringFlag{
			EnvVar: "AZURE_PUBLISH_SETTINGS_FILE",
			Name:   "azure-publish-settings-file",
			Usage:  "Azure publish settings file",
		},
		mcnflag.StringFlag{
			EnvVar: "AZURE_SIZE",
			Name:   "azure-size",
			Usage:  "Azure size",
			Value:  defaultSize,
		},
		mcnflag.IntFlag{
			Name:  "azure-ssh-port",
			Usage: "Azure SSH port",
			Value: defaultSSHPort,
		},
		mcnflag.StringFlag{
			EnvVar: "AZURE_SUBSCRIPTION_CERT",
			Name:   "azure-subscription-cert",
			Usage:  "Azure subscription cert",
		},
		mcnflag.StringFlag{
			EnvVar: "AZURE_SUBSCRIPTION_ID",
			Name:   "azure-subscription-id",
			Usage:  "Azure subscription ID",
		},
		mcnflag.StringFlag{
			Name:  "azure-username",
			Usage: "Azure username",
			Value: defaultSSHUsername,
		},
		mcnflag.StringFlag{
			Name:  "azure-os",
			Usage: "OS for the Azure VM (Windows|Linux)",
			Value: drivers.LINUX,
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	d := &Driver{
		DockerPort:            defaultDockerPort,
		DockerSwarmMasterPort: defaultSwarmMasterPort,
		Location:              defaultLocation,
		Size:                  defaultSize,
		BaseDriver: &drivers.BaseDriver{
			SSHPort:     defaultSSHPort,
			SSHUser:     defaultSSHUsername,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
	return d
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "ubuntu"
	}

	return d.SSHUser
}

func (d *Driver) GetOS() string {
	return d.OS
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "azure"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.SubscriptionID = flags.String("azure-subscription-id")

	cert := flags.String("azure-subscription-cert")
	publishSettings := flags.String("azure-publish-settings-file")
	image := flags.String("azure-image")
	username := flags.String("azure-username")
	OS := flags.String("azure-os")
	d.OS = strings.ToLower(OS)

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

	if d.OS == drivers.WINDOWS {
		image = defaultWindowsImage
	}

	if image == "" {
		d.Image = defaultLinuxImage
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
	d.DockerSwarmMasterPort = flags.Int("azure-docker-swarm-master-port")
	d.SSHPort = flags.Int("azure-ssh-port")
	d.SetSwarmConfigFromFlags(flags)

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	// check azure DNS to make sure name is available
	_, err = hostedservice.NewClient(client).CheckHostedServiceNameAvailability(d.MachineName)
	if err != nil {
		return err
	}

	return nil
}

//
// getServiceCertFingerprint returns the thumbprint of the certificate in a
// file
//
// Parameters:
//   certPath: path to the cert file
// Returns:
//   string: thumbprint string
//   error: error from reading the file or decoding cert
//
func getServiceCertFingerprint(certPath string) (string, error) {
	certData, readErr := ioutil.ReadFile(certPath)
	if readErr != nil {
		return "", readErr
	}

	block, rest := pem.Decode(certData)
	if block == nil {
		return "", errors.New(string(rest))
	}

	sha1sum := sha1.Sum(block.Bytes)
	fingerprint := fmt.Sprintf("%X", sha1sum)
	return fingerprint, nil
}

//
// configureForLinux sets up the VM role for Linux specific configuration
//
// Paramters:
//   role: role that needs to be updated with Linux configuration
//   dnsName: name of the machine that we are trying to create
//
// Returns:
//   error: errors from reading certs, getting thumbprint and adding
//   certificate to hostedservice
//
func (d *Driver) configureForLinux(role *virtualmachine.Role, dnsName string) error {
	// Get the Azure client
	client, err := d.getClient()
	if err != nil {
		return err
	}

	// Setup the image configuration
	vmutils.ConfigureDeploymentFromPlatformImage(
		role,
		d.Image,
		// XXX - need to query storage service to find the right
		// storage backend based on the location
		fmt.Sprintf("http://azurevmpp1.blob.core.windows.net/vhds/%s.vhd", dnsName),
		"")

	// Read the certificate
	data, err := ioutil.ReadFile(d.azureCertPath())
	if err != nil {
		return err
	}

	// Add the certificate to the hostedservice
	if _, err := hostedservice.NewClient(client).AddCertificate(dnsName, data, "pfx", ""); err != nil {
		return err
	}

	thumbPrint, err := getServiceCertFingerprint(d.azureCertPath())
	if err != nil {
		return err
	}

	vmutils.ConfigureForLinux(role, dnsName, d.SSHUser, d.UserPassword, thumbPrint)
	vmutils.ConfigureWithPublicSSH(role)

	role.UseCertAuth = true
	role.CertPath = d.azureCertPath()
	return nil
}

//
// configureForWindows sets up the VM role for Windows specific configuration
//
// Paramters:
//   role: role that needs to be updated with Windows configuration
//   dnsName: name of the machine that we are trying to create
//
// Returns:
//   None
//
func (d *Driver) configureForWindows(role *virtualmachine.Role, dnsName string) {
	vmutils.ConfigureDeploymentFromUserVMImage(
		role,
		d.Image)
	vmutils.ConfigureForWindows(role, dnsName, d.SSHUser, d.UserPassword, true, "")
	vmutils.ConfigureWithPublicSSH(role)
	vmutils.ConfigureWithPublicRDP(role)
	vmutils.ConfigureWithPublicPowerShell(role)
	vmutils.ConfigureWithExternalPort(role, "WinRMu", 5985, 5985,
		virtualmachine.InputEndpointProtocolTCP)
}

func (d *Driver) Create() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	log.Info("Creating Azure machine...")
	client, err := d.getClient()
	if err != nil {
		return err
	}

	dnsName := d.MachineName

	// 1. Create the hosted service
	if err := hostedservice.NewClient(client).CreateHostedService(hostedservice.CreateHostedServiceParameters{
		ServiceName: dnsName,
		Location:    d.Location,
		Label:       base64.StdEncoding.EncodeToString([]byte(dnsName))}); err != nil {
		return err
	}

	// 2. Generate certificates
	log.Debug("Generating certificate for Azure...")
	if err := d.generateCertForAzure(); err != nil {
		return err
	}

	// 3. Setup VM configuration for creation
	log.Debug("Setting up VM configuration...")
	var operationID management.OperationID

	role := vmutils.NewVMConfiguration(dnsName, d.Size)
	if d.OS == drivers.LINUX {
		err := d.configureForLinux(&role, dnsName)
		if err != nil {
			goto fail
		}

	} else {
		d.configureForWindows(&role, dnsName)
	}

	log.Debug("Authorizing docker ports...")
	if err := d.addDockerEndpoints(&role); err != nil {
		goto fail
	}

	// 4. Create the VM
	operationID, err = virtualmachine.NewClient(client).
		CreateDeployment(role, dnsName, virtualmachine.CreateDeploymentOptions{})
	if err != nil {
		goto fail
	}

	// 5. Wait for operation
	err = client.WaitForOperation(operationID, nil)
	if err != nil {
		goto fail
	}
	goto success

fail:
	hostedservice.NewClient(client).DeleteHostedService(dnsName, true)
	return err

success:
	return nil
}

//
// getClient returns a client for Azure API endpoint. Requires the
// publishsettings file to be already specified to the driver
//
// Parameters:
//   None
// Returns:
//   management.Client: client object to Azure API endpoint
//   error: errors from establishing new client
//
func (d *Driver) getClient() (management.Client, error) {
	client, err := management.ClientFromPublishSettingsFile(d.PublishSettingsFilePath, "")
	if err != nil {
		return management.NewAnonymousClient(), err
	}
	return client, nil
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

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

	client, err := d.getClient()
	if err != nil {
		return state.Error, err
	}

	dockerVM, err := virtualmachine.NewClient(client).GetDeployment(d.MachineName, d.MachineName)
	if err != nil {
		if strings.Contains(err.Error(), "Code: ResourceNotFound") {
			return state.Error, errors.New("Azure host was not found. Please check your Azure subscription.")
		}

		return state.Error, err
	}

	vmState := dockerVM.RoleInstanceList[0].PowerState
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

	client, err := d.getClient()
	if err != nil {
		return err
	}

	if _, err := virtualmachine.NewClient(client).StartRole(d.MachineName, d.MachineName, d.MachineName); err != nil {
		return err
	}

	d.IPAddress, err = d.GetIP()
	return err
}

func (d *Driver) Stop() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	if _, err := virtualmachine.NewClient(client).ShutdownRole(d.MachineName, d.MachineName, d.MachineName); err != nil {
		return err
	}

	d.IPAddress = ""
	return nil
}

func (d *Driver) Restart() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	if _, err := virtualmachine.NewClient(client).RestartRole(d.MachineName, d.MachineName, d.MachineName); err != nil {
		return err
	}

	d.IPAddress, err = d.GetIP()
	return err
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Remove() error {
	if err := d.setUserSubscription(); err != nil {
		return err
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	hostedClient := hostedservice.NewClient(client)
	if _, err := hostedClient.CheckHostedServiceNameAvailability(d.MachineName); err != nil {
		return err
	}

	_, err = hostedClient.DeleteHostedService(d.MachineName, true)
	return err
}

func (d *Driver) setUserSubscription() error {
	if d.PublishSettingsFilePath != "" {
		return ImportPublishSettingsFile(d.PublishSettingsFilePath)
	}
	return ImportPublishSettings(d.SubscriptionID, d.SubscriptionCert)
}

func (d *Driver) addDockerEndpoints(vmConfig *virtualmachine.Role) error {
	configSets := vmConfig.ConfigurationSets
	if len(configSets) == 0 {
		return errors.New("no configuration set")
	}
	for i := 0; i < len(configSets); i++ {
		if configSets[i].ConfigurationSetType != "NetworkConfiguration" {
			continue
		}
		ep := virtualmachine.InputEndpoint{
			Name:      "docker",
			Protocol:  "tcp",
			Port:      d.DockerPort,
			LocalPort: d.DockerPort,
		}
		if d.SwarmMaster {
			swarmEp := virtualmachine.InputEndpoint{
				Name:      "docker swarm",
				Protocol:  "tcp",
				Port:      d.DockerSwarmMasterPort,
				LocalPort: d.DockerSwarmMasterPort,
			}
			configSets[i].InputEndpoints = append(configSets[i].InputEndpoints, swarmEp)
			log.Debugf("added Docker swarm master endpoint (port %d) to configuration", d.DockerSwarmMasterPort)
		}
		configSets[i].InputEndpoints = append(configSets[i].InputEndpoints, ep)
		log.Debugf("added Docker endpoint (port %d) to configuration", d.DockerPort)
	}
	return nil
}

func (d *Driver) generateCertForAzure() error {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	cmd := exec.Command("openssl", "req", "-x509", "-key", d.GetSSHKeyPath(), "-nodes", "-days", "365", "-newkey", "rsa:2048", "-out", d.azureCertPath(), "-subj", "/C=AU/ST=Some-State/O=InternetWidgitsPtyLtd/CN=\\*")
	return cmd.Run()
}

func (d *Driver) azureCertPath() string {
	return d.ResolveStorePath("azure_cert.pem")
}

func (d *Driver) getHostname() string {
	return d.MachineName + ".cloudapp.net"
}
