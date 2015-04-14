package digitalocean

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"code.google.com/p/goauth2/oauth"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/digitalocean/godo"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	AccessToken       string
	DropletID         int
	DropletName       string
	Image             string
	MachineName       string
	IPAddress         string
	Region            string
	SSHKeyID          int
	SSHUser           string
	SSHPort           int
	Size              string
	IPv6              bool
	Backups           bool
	PrivateNetworking bool
	CaCertPath        string
	PrivateKeyPath    string
	DriverKeyPath     string
	SwarmMaster       bool
	SwarmHost         string
	SwarmDiscovery    string
	storePath         string
}

func init() {
	drivers.Register("digitalocean", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_ACCESS_TOKEN",
			Name:   "digitalocean-access-token",
			Usage:  "Digital Ocean access token",
		},
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_IMAGE",
			Name:   "digitalocean-image",
			Usage:  "Digital Ocean Image",
			Value:  "ubuntu-14-04-x64",
		},
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_REGION",
			Name:   "digitalocean-region",
			Usage:  "Digital Ocean region",
			Value:  "nyc3",
		},
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_SIZE",
			Name:   "digitalocean-size",
			Usage:  "Digital Ocean size",
			Value:  "512mb",
		},
		cli.BoolFlag{
			EnvVar: "DIGITALOCEAN_IPV6",
			Name:   "digitalocean-ipv6",
			Usage:  "enable ipv6 for droplet",
		},
		cli.BoolFlag{
			EnvVar: "DIGITALOCEAN_PRIVATE_NETWORKING",
			Name:   "digitalocean-private-networking",
			Usage:  "enable private networking for droplet",
		},
		cli.BoolFlag{
			EnvVar: "DIGITALOCEAN_BACKUPS",
			Name:   "digitalocean-backups",
			Usage:  "enable backups for droplet",
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

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Remote
}

func (d *Driver) DriverName() string {
	return "digitalocean"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessToken = flags.String("digitalocean-access-token")
	d.Image = flags.String("digitalocean-image")
	d.Region = flags.String("digitalocean-region")
	d.Size = flags.String("digitalocean-size")
	d.IPv6 = flags.Bool("digitalocean-ipv6")
	d.PrivateNetworking = flags.Bool("digitalocean-private-networking")
	d.Backups = flags.Bool("digitalocean-backups")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "root"
	d.SSHPort = 22

	if d.AccessToken == "" {
		return fmt.Errorf("digitalocean driver requires the --digitalocean-access-token option")
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	client := d.getClient()
	regions, _, err := client.Regions.List(nil)
	if err != nil {
		return err
	}
	for _, region := range regions {
		if region.Slug == d.Region {
			return nil
		}
	}

	return fmt.Errorf("digitalocean requires a valid region")
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}

	d.SSHKeyID = key.ID

	log.Infof("Creating Digital Ocean droplet...")

	client := d.getClient()

	createRequest := &godo.DropletCreateRequest{
		Image:             d.Image,
		Name:              d.MachineName,
		Region:            d.Region,
		Size:              d.Size,
		IPv6:              d.IPv6,
		PrivateNetworking: d.PrivateNetworking,
		Backups:           d.Backups,
		SSHKeys:           []interface{}{d.SSHKeyID},
	}

	newDroplet, _, err := client.Droplets.Create(createRequest)
	if err != nil {
		return err
	}

	d.DropletID = newDroplet.Droplet.ID

	for {
		newDroplet, _, err = client.Droplets.Get(d.DropletID)
		if err != nil {
			return err
		}
		for _, network := range newDroplet.Droplet.Networks.V4 {
			if network.Type == "public" {
				d.IPAddress = network.IPAddress
			}
		}

		if d.IPAddress != "" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Debugf("Created droplet ID %d, IP address %s",
		newDroplet.Droplet.ID,
		d.IPAddress)

	return nil
}

func (d *Driver) createSSHKey() (*godo.Key, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	createRequest := &godo.KeyCreateRequest{
		Name:      d.MachineName,
		PublicKey: string(publicKey),
	}

	key, _, err := d.getClient().Keys.Create(createRequest)
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
	droplet, _, err := d.getClient().Droplets.Get(d.DropletID)
	if err != nil {
		return state.Error, err
	}
	switch droplet.Droplet.Status {
	case "new":
		return state.Starting, nil
	case "active":
		return state.Running, nil
	case "off":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	_, _, err := d.getClient().DropletActions.PowerOn(d.DropletID)
	return err
}

func (d *Driver) Stop() error {
	_, _, err := d.getClient().DropletActions.Shutdown(d.DropletID)
	return err
}

func (d *Driver) Remove() error {
	client := d.getClient()
	if resp, err := client.Keys.DeleteByID(d.SSHKeyID); err != nil {
		if resp.StatusCode == 404 {
			log.Infof("Digital Ocean SSH key doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}
	if resp, err := client.Droplets.Delete(d.DropletID); err != nil {
		if resp.StatusCode == 404 {
			log.Infof("Digital Ocean droplet doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}
	return nil
}

func (d *Driver) Restart() error {
	_, _, err := d.getClient().DropletActions.Reboot(d.DropletID)
	return err
}

func (d *Driver) Kill() error {
	_, _, err := d.getClient().DropletActions.PowerOff(d.DropletID)
	return err
}

func (d *Driver) getClient() *godo.Client {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: d.AccessToken},
	}

	return godo.NewClient(t.Client())
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}
