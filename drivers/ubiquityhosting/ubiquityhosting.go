package ubiquityhosting

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/ubiquityhosting/GoUbi"
	"io/ioutil"
	"path/filepath"
	"time"
)

type Driver struct {
	ServiceID int
	ClientID  int
	Username  string
	Token     string
	ZoneID    int
	FlavorID  int
	ImageID   int
	Userdata  string

	MachineName    string
	IPAddress      string
	SSHKeyID       int
	SSHUser        string
	SSHPort        int
	CaCertPath     string
	PrivateKeyPath string
	DriverKeyPath  string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
	storePath      string
}

func init() {
	drivers.Register("ubiquity", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})

}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "UBIQUITY_CLIENT_ID",
			Name:   "ubiquity-client-id",
			Usage:  "Ubiquity client ID for account authentication",
		},
		cli.StringFlag{
			EnvVar: "UBIQUITY_API_USERNAME",
			Name:   "ubiquity-api-username",
			Usage:  "Ubiquity username for API authentication",
		},
		cli.StringFlag{
			EnvVar: "UBIQUITY_API_TOKEN",
			Name:   "ubiquity-api-token",
			Usage:  "Ubiquity API token for authentication",
		},
		cli.StringFlag{
			EnvVar: "UBIQUITY_ZONE_ID",
			Name:   "ubiquity-zone-id",
			Usage:  "Ubiquity zone location for VM creation",
			Value:  "7",
		},
		cli.StringFlag{
			EnvVar: "UBIQUITY_FLAVOR_ID",
			Name:   "ubiquity-flavor-id",
			Usage:  "Ubiquity VM size details for VM creation",
			Value:  "1",
		},
		cli.StringFlag{
			EnvVar: "UBIQUITY_IMAGE_ID",
			Name:   "ubiquity-image-id",
			Usage:  "Ubiquity VM image for VM creation",
			Value:  "18",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
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
		d.SSHUser = "root"
	}

	return d.SSHUser
}

func (d *Driver) DriverName() string {
	return "ubiquity"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ClientID = flags.Int("ubiquity-client-id")
	d.Username = flags.String("ubiquity-api-username")
	d.Token = flags.String("ubiquity-api-token")
	d.ZoneID = flags.Int("ubiquity-zone-id")
	d.FlavorID = flags.Int("ubiquity-flavor-id")
	d.ImageID = flags.Int("ubiquity-image-id")
	d.Userdata = flags.String("ubiquity-user-data")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "root"
	d.SSHPort = 22

	if d.Token == "" {
		return fmt.Errorf("ubiquity driver requires the -ubiquity-api-token option")
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}

	d.SSHKeyID = key
	log.Debugf("Created SSH Key ID: %d", key)

	log.Infof("Creating Ubiquity instance, please wait...")

	createRequest := &goubi.CreateVMParams{
		Hostname:      d.MachineName,
		ImageID:       d.ImageID,
		FlavorID:      d.FlavorID,
		ZoneID:        d.ZoneID,
		KeyID:         d.SSHKeyID,
		DockerMachine: true,
	}

	client := d.getClient()

	instance, err := client.Cloud.Create(createRequest)
	if err != nil {
		return err
	}
	log.Debugf("Instance service ID: %d", instance.ServiceID)
	d.ServiceID = instance.ServiceID

	for {
		details, err := client.Cloud.Get(d.ServiceID)
		if err != nil {
			log.Debugf("Waiting for VM creation... (Error: %v)", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Debug("Instance created")
		d.IPAddress = details.MainIPaddress

		if d.IPAddress != "" {
			break
		}
	}
	log.Infof("Initializing instance on IP address: %s", d.IPAddress)
	log.Debugf("Created instance ID %d, IP address %s",
		d.ServiceID,
		d.IPAddress)

	return nil
}

func (d *Driver) createSSHKey() (int, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return 0, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return 0, err
	}

	createRequest := &goubi.AddKeyParams{
		KeyName: d.MachineName,
		PubKey:  string(publicKey),
	}

	key, err := d.getClient().Cloud.AddKey(createRequest)
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
	instance, err := d.getClient().Cloud.Get(d.ServiceID)
	if err != nil {
		return state.Error, err
	}
	switch instance.State {
	case "online":
		return state.Running, nil
	case "offline":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	_, err := d.getClient().Cloud.Start(d.ServiceID)
	return err
}

func (d *Driver) Stop() error {
	_, err := d.getClient().Cloud.Stop(d.ServiceID)
	return err
}

func (d *Driver) Remove() error {
	client := d.getClient()

	if _, err := client.Cloud.RemoveKey(d.SSHKeyID); err != nil {
		log.Infof(err.Error())
	}
	if _, err := client.Cloud.Destroy(d.ServiceID); err != nil {
		log.Infof(err.Error())
	}
	return nil
}

func (d *Driver) Restart() error {
	_, err := d.getClient().Cloud.Reboot(d.ServiceID)
	return err
}

func (d *Driver) Kill() error {
	_, err := d.getClient().Cloud.Stop(d.ServiceID)
	return err
}

func (d *Driver) getClient() *goubi.UbiServices {
	return goubi.NewUbiClient(d.ClientID, d.Username, d.Token, true)
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}
