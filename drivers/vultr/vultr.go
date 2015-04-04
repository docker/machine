package vultr

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	vultr "github.com/JamesClonk/vultr/lib"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	dockerConfigDir = "/etc/docker"
)

type Driver struct {
	APIKey            string
	MachineID         string
	MachineName       string
	PublicIP          string
	PrivateIP         string
	OSID              int
	RegionID          int
	PlanID            int
	SSHKeyID          string
	SSHUser           string
	SSHPort           int
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
	drivers.Register("vultr", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "VULTR_API_KEY",
			Name:   "vultr-api-key",
			Usage:  "Vultr API key",
		},
		cli.IntFlag{
			EnvVar: "VULTR_OS",
			Name:   "vultr-os-id",
			Usage:  "Vultr operating system ID (OSID). Default: 160 (Ubuntu 14.04 x64)",
			Value:  160,
		},
		cli.IntFlag{
			EnvVar: "VULTR_REGION",
			Name:   "vultr-region-id",
			Usage:  "Vultr region ID (DCID). Default: 1 (New Jersey)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "VULTR_PLAN",
			Name:   "vultr-plan-id",
			Usage:  "Vultr plan ID (VPSPLANID). Default: 29 (768 MB RAM, 15 GB SSD, 1.00 TB BW)",
			Value:  29,
		},
		cli.BoolFlag{
			EnvVar: "VULTR__IPV6",
			Name:   "digitalocean-ipv6",
			Usage:  "enable ipv6 for droplet",
		},
		cli.BoolFlag{
			EnvVar: "VULTR_PRIVATE_NETWORKING",
			Name:   "digitalocean-private-networking",
			Usage:  "enable private networking for this virtual machine",
		},
		cli.BoolFlag{
			EnvVar: "VULTR_BACKUPS",
			Name:   "vultr-backups",
			Usage:  "enable automatic backups for this virtual machine",
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
	return "vultr"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIKey = flags.String("vultr-api-key")
	d.OSID = flags.Int("vultr-os-id")
	d.RegionID = flags.Int("vultr-region-id")
	d.PlanID = flags.Int("vultr-plan-id")
	d.IPv6 = flags.Bool("vultr-ipv6")
	d.PrivateNetworking = flags.Bool("vultr-private-networking")
	d.Backups = flags.Bool("vultr-backups")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "root"
	d.SSHPort = 22

	if d.APIKey == "" {
		return fmt.Errorf("vultr driver requires the --vultr-api-key option")
	}
	return nil
}

func (d *Driver) PreCreateCheck() error {
	client := d.getClient()
	regions, _, err := client.GetRegions()
	if err != nil {
		return err
	}
	for _, region := range regions {
		if region.ID == d.RegionID {
			return nil
		}
	}

	return fmt.Errorf("vultr requires a valid region")
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}
	d.SSHKeyID = key.ID

	log.Infof("Creating Vultr virtual machine...")
	client := d.getClient()
	machine, err := client.CreateServer(
		d.MachineName,
		d.RegionID,
		d.PlanID,
		d.OSID,
		&vultr.ServerOptions{
			SSHKey:            d.SSHKeyID,
			IPV6:              d.IPv6,
			PrivateNetworking: d.PrivateNetworking,
			AutoBackups:       d.Backups,
		})
	if err != nil {
		return err
	}
	d.MachineID = machine.ID

	for {
		machine, err = client.GetServer(d.MachineID)
		if err != nil {
			return err
		}
		d.PublicIP = machine.MainIP
		d.PrivateIP = machine.InternalIP

		if d.PublicIP != "" && d.PublicIP != "0" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Debugf("Created virtual machine ID %s, IP address %s, Private IP address %s",
		d.InstanceId,
		d.IPAddress,
		d.PrivateIPAddress,
	)

	log.Infof("Waiting for SSH...")
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.PublicIP, 22)); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createSSHKey() (*vultr.SSHKey, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	key, err := d.getClient().CreateSSHKey(d.MachineName, string(publicKey))
	if err != nil {
		return &key, err
	}
	return &key, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.PublicIP == "" || d.PublicIP == "0" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.PublicIP, nil
}

func (d *Driver) GetState() (state.State, error) {
	machine, err := d.getClient().GetServer(d.MachineID)
	if err != nil {
		return state.Error, err
	}
	switch machine.Status {
	case "pending":
		return state.Starting, nil
	case "active":
		switch machine.PowerStatus {
		case "running":
			return state.Running, nil
		case "stopped":
			return state.Stopped, nil
		}
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	return d.getClient().StartServer(d.MachineID)
}

func (d *Driver) Stop() error {
	return d.getClient().HaltServer(d.MachineID)
}

func (d *Driver) Remove() error {
	client := d.getClient()
	if err := client.DeleteServer(d.MachineID); err != nil {
		return err
	}
	if err := client.DeleteSSHKey(d.SSHKeyID); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	return d.getClient().RebootServer(d.MachineID)
}

func (d *Driver) Kill() error {
	return d.getClient().HaltServer(d.MachineID)
}

func (d *Driver) GetDockerConfigDir() string {
	return dockerConfigDir
}

func (d *Driver) getClient() *vultr.Client {
	return vultr.NewClient(d.APIKey, nil)
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}
