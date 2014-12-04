package azure

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	azure "github.com/MSOpenTech/azure-sdk-for-go"
	"github.com/MSOpenTech/azure-sdk-for-go/clients/vmClient"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
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

type CreateFlags struct {
	SubscriptionID          *string
	SubscriptionCert        *string
	PublishSettingsFilePath *string
	Name                    *string
	Location                *string
	Size                    *string
	UserName                *string
	UserPassword            *string
	Image                   *string
	SSHPort                 *string
	DockerPort              *string
}

func init() {
	drivers.Register("azure", &drivers.RegisteredDriver{
		New:                 NewDriver,
		RegisterCreateFlags: RegisterCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
	createFlags := new(CreateFlags)
	createFlags.SubscriptionID = cmd.String(
		[]string{"-azure-subscription-id"},
		"",
		"Azure subscription ID",
	)
	createFlags.SubscriptionCert = cmd.String(
		[]string{"-azure-subscription-cert"},
		"",
		"Azure subscription cert",
	)
	createFlags.PublishSettingsFilePath = cmd.String(
		[]string{"-azure-publish-settings-file"},
		"",
		"Azure publish settings file",
	)
	createFlags.Location = cmd.String(
		[]string{"-azure-location"},
		"West US",
		"Azure location",
	)
	createFlags.Size = cmd.String(
		[]string{"-azure-size"},
		"Small",
		"Azure size",
	)
	createFlags.Name = cmd.String(
		[]string{"-azure-name"},
		"",
		"Azure cloud service name",
	)
	createFlags.UserName = cmd.String(
		[]string{"-azure-username"},
		"tcuser",
		"Azure username",
	)
	createFlags.UserPassword = cmd.String(
		[]string{"-azure-password"},
		"",
		"Azure user password",
	)
	createFlags.Image = cmd.String(
		[]string{"-azure-image"},
		"",
		"Azure image name. Default is Ubuntu 14.04 LTS x64",
	)
	createFlags.SSHPort = cmd.String(
		[]string{"-azure-ssh"},
		"22",
		"Azure ssh port",
	)
	createFlags.DockerPort = cmd.String(
		[]string{"-azure-docker-port"},
		"4243",
		"Azure docker port",
	)
	return createFlags
}

func NewDriver(storePath string) (drivers.Driver, error) {
	driver := &Driver{storePath: storePath}
	return driver, nil
}

func (driver *Driver) DriverName() string {
	return "azure"
}

func (driver *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
	flags := flagsInterface.(*CreateFlags)
	driver.SubscriptionID = *flags.SubscriptionID

	if *flags.SubscriptionCert != "" {
		if _, err := os.Stat(*flags.SubscriptionCert); os.IsNotExist(err) {
			return err
		}
		driver.SubscriptionCert = *flags.SubscriptionCert
	}

	if *flags.PublishSettingsFilePath != "" {
		if _, err := os.Stat(*flags.PublishSettingsFilePath); os.IsNotExist(err) {
			return err
		}
		driver.PublishSettingsFilePath = *flags.PublishSettingsFilePath
	}

	if (driver.SubscriptionID == "" || driver.SubscriptionCert == "") && driver.PublishSettingsFilePath == "" {
		return fmt.Errorf("Please specify azure subscription params using options: --azure-subscription-id and --azure-subscription-cert or --azure-publish-settings-file")
	}

	if *flags.Name == "" {
		driver.Name = generateVMName()
	} else {
		driver.Name = *flags.Name
	}

	if *flags.Image == "" {
		driver.Image = "b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB"
	} else {
		driver.Image = *flags.Image
	}

	driver.Location = *flags.Location
	driver.Size = *flags.Size

	if strings.ToLower(*flags.UserName) == "docker" {
		return fmt.Errorf("'docker' is not valid user name for docker host. Please specify another user name")
	}

	driver.UserName = *flags.UserName
	driver.UserPassword = *flags.UserPassword

	dockerPort, err := strconv.Atoi(*flags.DockerPort)
	if err != nil {
		return err
	}
	driver.DockerPort = dockerPort

	sshPort, err := strconv.Atoi(*flags.SSHPort)
	if err != nil {
		return err
	}
	driver.SSHPort = sshPort

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

	_, err = vmClient.GetVMDeployment(driver.Name, driver.Name)
	if err != nil {
		if strings.Contains(err.Error(), "Code: ResourceNotFound") {
			return nil
		}

		return err
	}

	err = vmClient.DeleteVMDeployment(driver.Name, driver.Name)
	if err != nil {
		return err
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
	return path.Join(driver.storePath, "id_rsa")
}

func (driver *Driver) publicSSHKeyPath() string {
	return driver.sshKeyPath() + ".pub"
}

func (driver *Driver) azureCertPath() string {
	return path.Join(driver.storePath, "azure_cert.pem")
}
