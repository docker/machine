package packet

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/packethost/packngo"
)

const (
	dockerConfigDir = "/etc/docker"
	consumerToken   = "24e70949af5ecd17fe8e867b335fc88e7de8bd4ad617c0403d8769a376ddea72"
)

func init() {
	drivers.Register("packet", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "PACKET_API_KEY",
			Name:   "packet-api-key",
			Usage:  "Packet api key",
		},
		cli.StringFlag{
			EnvVar: "PACKET_PROJECT_ID",
			Name:   "packet-project-id",
			Usage:  "Packet Project Id",
		},
		cli.StringFlag{
			EnvVar: "PACKET_OS",
			Name:   "packet-os",
			Usage:  fmt.Sprintf("Packet OS, possible values are: %v", strings.Join(getOsFlavors(), ", ")),
			Value:  "ubuntu_14_04",
		},
		cli.StringFlag{
			EnvVar: "PACKET_FACILITY_CODE",
			Name:   "packet-facility-code",
			Usage:  "Packet facility code",
			Value:  "ewr1",
		},
		cli.StringFlag{
			EnvVar: "PACKET_PLAN",
			Name:   "packet-plan",
			Usage:  "Packet Server Plan",
			Value:  "baremetal_1",
		},
		cli.StringFlag{
			EnvVar: "PACKET_BILLING_CYCLE",
			Name:   "packet-billing-cycle",
			Usage:  "Packet billing cycle, hourly or monthly",
			Value:  "hourly",
		},
	}
}

func getOsFlavors() []string {
	return []string{"ubuntu_14_04"}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, StorePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

type Driver struct {
	ConsumerToken   string
	ApiKey          string
	ProjectID       string
	Plan            string
	Facility        string
	OperatingSystem string
	BillingCycle    string
	MachineName     string
	DeviceID        string
	UserData        string
	Tags            []string
	IPAddress       string
	CaCertPath      string
	PrivateKeyPath  string
	SSHKeyID        string
	SSHUser         string
	SSHPort         int
	StorePath       string
}

func (d *Driver) DriverName() string {
	return "packet"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ConsumerToken = consumerToken

	if strings.Contains(flags.String("packet-os"), "coreos") {
		d.SSHUser = "core"
	} else {
		d.SSHUser = "root"
	}

	d.ApiKey = flags.String("packet-api-key")
	d.ProjectID = flags.String("packet-project-id")
	d.OperatingSystem = flags.String("packet-os")
	d.Facility = flags.String("packet-facility-code")
	d.Plan = flags.String("packet-plan")
	d.BillingCycle = flags.String("packet-billing-cycle")

	if d.ApiKey == "" {
		return fmt.Errorf("packet driver requires the --packet-api-key option")
	}
	if d.ProjectID == "" {
		return fmt.Errorf("packet driver requires the --packet-project-id option")
	}

	return nil
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
	return filepath.Join(d.StorePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "root"
	}

	return d.SSHUser
}

func (d *Driver) PreCreateCheck() error {
	if !stringInSlice(d.OperatingSystem, getOsFlavors()) {
		return fmt.Errorf("specified --packet-os not one of %v", strings.Join(getOsFlavors(), ", "))
	}

	client := d.getClient()
	facilities, _, err := client.Facilities.List()
	if err != nil {
		return err
	}
	for _, facility := range facilities {
		if facility.Code == d.Facility {
			return nil
		}
	}

	return fmt.Errorf("packet requires a valid facility")
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}

	d.SSHKeyID = key.ID

	client := d.getClient()
	createRequest := &packngo.DeviceCreateRequest{
		HostName:     d.MachineName,
		Plan:         d.Plan,
		Facility:     d.Facility,
		OS:           d.OperatingSystem,
		BillingCycle: d.BillingCycle,
		ProjectID:    d.ProjectID,
		UserData:     d.UserData,
		Tags:         d.Tags,
	}

	log.Infof("Provisioning Packet server...")
	newDevice, _, err := client.Devices.Create(createRequest)
	if err != nil {
		return err
	}
	t0 := time.Now()

	d.DeviceID = newDevice.ID

	for {
		newDevice, _, err = client.Devices.Get(d.DeviceID)
		if err != nil {
			return err
		}

		for _, ip := range newDevice.Network {
			if ip.Public && ip.Family == 4 {
				d.IPAddress = ip.Address
			}
		}

		if d.IPAddress != "" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Infof("Created device ID %s, IP address %s",
		newDevice.ID,
		d.IPAddress)

	log.Infof("Waiting for Provisioning...")
	stage := float32(0)
	for {
		newDevice, _, err = client.Devices.Get(d.DeviceID)
		if err != nil {
			return err
		}
		if newDevice.State == "provisioning" && stage != newDevice.ProvisionPer {
			stage = newDevice.ProvisionPer
			log.Debug("Provisioning %v%% complete", newDevice.ProvisionPer)
		}
		if newDevice.State == "active" {
			log.Debug("Device State: %s", newDevice.State)
			break
		}
		time.Sleep(10 * time.Second)
	}

	log.Debug("Provision time: %v.\n", time.Since(t0))

	log.Debug("Waiting for SSH...")
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, 22)); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createSSHKey() (*packngo.SSHKey, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	createRequest := &packngo.SSHKeyCreateRequest{
		Label: fmt.Sprintf("docker machine: %s", d.MachineName),
		Key:   string(publicKey),
	}

	key, _, err := d.getClient().SSHKeys.Create(createRequest)
	if err != nil {
		return key, err
	}

	return key, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	device, _, err := d.getClient().Devices.Get(d.DeviceID)
	if err != nil {
		return state.Error, err
	}

	switch device.State {
	case "provisioning":
	case "powering_on":
		return state.Starting, nil
	case "active":
		return state.Running, nil
	case "powering_off":
		return state.Stopping, nil
	case "inactive":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	_, err := d.getClient().Devices.PowerOn(d.DeviceID)
	return err
}

func (d *Driver) Stop() error {
	_, err := d.getClient().Devices.PowerOff(d.DeviceID)
	return err
}

func (d *Driver) Remove() error {
	client := d.getClient()

	if _, err := client.SSHKeys.Delete(d.SSHKeyID); err != nil {
		return err
	}

	if _, err := client.Devices.Delete(d.DeviceID); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	_, err := d.getClient().Devices.Reboot(d.DeviceID)
	return err
}

func (d *Driver) Kill() error {
	_, err := d.getClient().Devices.PowerOff(d.DeviceID)
	return err
}

func (d *Driver) GetDockerConfigDir() string {
	return dockerConfigDir
}

func (d *Driver) getClient() *packngo.Client {
	return packngo.NewClient(d.ConsumerToken, d.ApiKey)
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.StorePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
