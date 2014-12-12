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

type Driver struct {
	SubscriptionID          string
	SubscriptionCert        string
	PublishSettingsFilePath string
	Name                    string
	Location                string
	Size                    string
	UserName                string
	UserPassword            string
	Image                   string
	SSHPort                 int
	DockerPort              int
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
			Value: 4243,
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
			Name:  "azure-name",
			Usage: "Azure cloud service name",
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
			Name:  "azure-ssh",
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
			Value: "tcuser",
		},
	}
}

func NewDriver(storePath string) (drivers.Driver, error) {
	driver := &Driver{storePath: storePath}
	return driver, nil
}

func (driver *Driver) DriverName() string {
	return "azure"
}

func (driver *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	driver.SubscriptionID = flags.String("azure-subscription-id")

	cert := flags.String("azure-subscription-cert")
	publishSettings := flags.String("azure-publish-settings-file")
	name := flags.String("azure-name")
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

	if name == "" {
		driver.Name = generateVMName()
	} else {
		driver.Name = name
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
	driver.SSHPort = flags.Int("azure-docker-port")

	return nil
}

func (driver *Driver) Create() error {
	if err := driver.setUserSubscription(); err != nil {
		return err
	}

	log.Infof("Creating Azure host...")
	vmConfig, err := vmClient.CreateAzureVMConfiguration(driver.Name, driver.Size, driver.Image, driver.Location)
	if err != nil {
		return err
	}

	if err := driver.generateCertForAzure(); err != nil {
		return err
	}

	vmConfig, err = vmClient.AddAzureLinuxProvisioningConfig(vmConfig, driver.UserName, driver.UserPassword, driver.azureCertPath(), driver.SSHPort)
	if err != nil {
		return err
	}

	vmConfig, err = vmClient.SetAzureDockerVMExtension(vmConfig, driver.DockerPort, "0.4")
	if err != nil {
		return err
	}

	if err := vmClient.CreateAzureVM(vmConfig, driver.Name, driver.Location); err != nil {
		return err
	}

	if err := driver.waitForSSH(); err != nil {
		return err
	}

	if err := driver.waitForDocker(); err != nil {
		return err
	}

	if err := driver.hackForIdentityAuth(); err != nil {
		return err
	}

	return nil
}

func (driver *Driver) hackForIdentityAuth() error {

	log.Debugf("HACK: Downloading version of Docker with identity auth...")

	/* We need to add retries to every SSH call we make, because Azure has some weird networking bug:
	sometimes when it comes to communication between VMs or with Azure itself, Azure API throws an error.
	So when we are running remote commands via SSH, sometimes they fail for no reason.
	This issue is fixed by repeating SSH calls few times before throwing an error.
	*/
	numberOfRetries := 3
	if err := driver.runSSHCommand("sudo stop docker", numberOfRetries); err != nil {
		return err
	}

	if err := driver.runSSHCommand("sudo bash -c \"curl -sS https://bfirsh.s3.amazonaws.com/docker/docker-1.3.1-dev-identity-auth > /usr/bin/docker\"", numberOfRetries); err != nil {
		return err
	}

	log.Debugf("Updating /etc/default/docker to use identity auth...")

	cmdString := fmt.Sprintf(`sudo bash -c 'cat <<EOF > /etc/default/docker
export DOCKER_OPTS="--auth=identity --host=tcp://0.0.0.0:%v"
EOF'`, driver.DockerPort)
	if err := driver.runSSHCommand(cmdString, numberOfRetries); err != nil {
		return err
	}

	log.Debugf("Adding key to authorized-keys.d...")

	if err := driver.addPublicKeyToAuthorizedHosts("/tmp/.docker/authorized-keys.d", numberOfRetries); err != nil {
		return err
	}

	if err := driver.runSSHCommand("sudo cp -a /tmp/.docker/ /", numberOfRetries); err != nil {
		return err
	}

	if err := driver.runSSHCommand("rm -r /tmp/.docker/", numberOfRetries); err != nil {
		return err
	}

	if err := driver.runSSHCommand("sudo start docker", numberOfRetries); err != nil {
		return err
	}

	return nil
}

func (driver *Driver) addPublicKeyToAuthorizedHosts(authorizedKeysPath string, retries int) error {
	if err := drivers.AddPublicKeyToAuthorizedHosts(driver, authorizedKeysPath); err != nil {
		if err.Error() == "exit status 255" {
			if retries == 0 {
				return err
			}
			return driver.addPublicKeyToAuthorizedHosts(authorizedKeysPath, retries-1)
		}

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
	url := fmt.Sprintf("tcp://%s.cloudapp.net:%v", driver.Name, driver.DockerPort)
	return url, nil
}

func (driver *Driver) GetIP() (string, error) {
	return fmt.Sprintf("%s.cloudapp.net", driver.Name), nil
}

func (driver *Driver) GetState() (state.State, error) {
	err := driver.setUserSubscription()
	if err != nil {
		return state.Error, err
	}

	dockerVM, err := vmClient.GetVMDeployment(driver.Name, driver.Name)
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

	err = vmClient.StartRole(driver.Name, driver.Name, driver.Name)
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
	err = vmClient.ShutdownRole(driver.Name, driver.Name, driver.Name)
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
	available, _, err := vmClient.CheckHostedServiceNameAvailability(driver.Name)
	if err != nil {
		return err
	}
	if available {
		return nil
	}
	err = vmClient.DeleteHostedService(driver.Name)
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
	err = vmClient.RestartRole(driver.Name, driver.Name, driver.Name)
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
	err = vmClient.ShutdownRole(driver.Name, driver.Name, driver.Name)
	if err != nil {
		return err
	}
	return nil
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

	return ssh.GetSSHCommand(driver.Name+".cloudapp.net", driver.SSHPort, driver.UserName, driver.sshKeyPath(), args...), nil
}

func (driver *Driver) Upgrade() error {
	return nil
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

func (driver *Driver) waitForSSH() error {
	log.Infof("Waiting for SSH...")
	err := ssh.WaitForTCP(fmt.Sprintf("%s:%v", driver.Name+".cloudapp.net", driver.SSHPort))
	if err != nil {
		return err
	}

	return nil
}

func (driver *Driver) waitForDocker() error {
	log.Infof("Waiting for docker daemon on host to be available...")
	maxRepeats := 48
	url := fmt.Sprintf("%s:%v", driver.Name+".cloudapp.net", driver.DockerPort)
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
