package vultr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/ChimeraCoder/tokenbucket"
	vultr "github.com/JamesClonk/vultr/lib"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

var vultrTokenBucket = tokenbucket.NewBucket(1*time.Second, 1)

type Driver struct {
	APIKey            string
	MachineID         string
	MachineName       string
	IPAddress         string
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
	scriptID          int
	bucket            *tokenbucket.Bucket
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
			Usage:  "Vultr operating system ID. Default: 160 (ubuntu-14-04-x64)",
			Value:  160,
		},
		cli.IntFlag{
			EnvVar: "VULTR_REGION",
			Name:   "vultr-region-id",
			Usage:  "Vultr region ID. Default: 1 (New Jersey)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "VULTR_PLAN",
			Name:   "vultr-plan-id",
			Usage:  "Vultr plan ID. Default: 29 (768 MB RAM)",
			Value:  29,
		},
		cli.BoolFlag{
			EnvVar: "VULTR_IPV6",
			Name:   "vultr-ipv6",
			Usage:  "Enable IPv6 for VPS",
		},
		cli.BoolFlag{
			EnvVar: "VULTR_PRIVATE_NETWORKING",
			Name:   "vultr-private-networking",
			Usage:  "Enable private networking for VPS",
		},
		cli.BoolFlag{
			EnvVar: "VULTR_BACKUPS",
			Name:   "vultr-backups",
			Usage:  "Enable automatic backups for VPS",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	d := &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey, bucket: nil}
	// throttle API queries according to 1/s rate-limit
	d.bucket = vultrTokenBucket
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
		d.SSHUser = "root"
	}

	return d.SSHUser
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

	log.Info("Validating Vultr VPS parameters...")

	if err := d.validateRegion(); err != nil {
		return err
	}

	if err := d.validatePlan(); err != nil {
		return err
	}

	if err := d.validateApiCredentials(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Create() error {
	var userdata string
	log.Debug("Generating SSH key...")
	key, err := d.createSSHKey()
	if err != nil {
		return err
	}
	d.SSHKeyID = key.ID

	log.Info("Creating Vultr VPS...")
	// RancherOS iPXE boot prep
	if d.OSID == 159 {
		log.Info("Deploying RancherOS")
		d.SSHUser = "rancher"
		if err := d.createBootScript(); err != nil {
			return err
		}
		log.Debugf("Uploaded iPXE boot script ID %d", d.scriptID)
		userdata, err = d.getCloudConfig()
		if err != nil {
			return err
		}
		log.Debugf("Using the following cloud-config file:")
		log.Debugf("%s", userdata)
	}

	client := d.getClient()
	<-d.bucket.SpendToken(1)
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
			Script:            d.scriptID,
			UserData:          userdata,
		})
	if err != nil {
		return err
	}
	d.MachineID = machine.ID

	log.Info("Waiting for IP address to become available...")
	for {
		<-d.bucket.SpendToken(3)
		machine, err = client.GetServer(d.MachineID)
		if err != nil {
			return err
		}
		d.IPAddress = machine.MainIP
		d.PrivateIP = machine.InternalIP

		if d.IPAddress != "" && d.IPAddress != "0" {
			break
		}
		log.Debug("IP address not yet available")
	}

	if d.PrivateIP == "0" {
		d.PrivateIP = ""
	}

	log.Infof("Created Vultr VPS ID: %s, Public IP: %s, Private IP: %s",
		d.MachineID,
		d.IPAddress,
		d.PrivateIP,
	)

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

	<-d.bucket.SpendToken(1)
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
	if d.IPAddress == "" || d.IPAddress == "0" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	<-d.bucket.SpendToken(1)
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
	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Running || vmState == state.Starting {
		log.Infof("Host is already running or starting")
		return nil
	}
	log.Debugf("starting %s", d.MachineName)
	<-d.bucket.SpendToken(1)
	return d.getClient().StartServer(d.MachineID)
}

func (d *Driver) Stop() error {
	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Stopped {
		log.Infof("Host is already stopped")
		return nil
	}
	log.Debugf("stopping %s", d.MachineName)
	<-d.bucket.SpendToken(1)
	return d.getClient().HaltServer(d.MachineID)
}

func (d *Driver) Remove() error {
	client := d.getClient()
	log.Debugf("removing %s", d.MachineName)
	<-d.bucket.SpendToken(1)
	if err := client.DeleteServer(d.MachineID); err != nil {
		if strings.Contains(err.Error(), "Invalid server") {
			log.Infof("VPS doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}
	<-d.bucket.SpendToken(1)
	if err := client.DeleteSSHKey(d.SSHKeyID); err != nil {
		if strings.Contains(err.Error(), "Invalid SSH Key") {
			log.Infof("SSH key doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}
	if d.scriptID != 0 {
		<-d.bucket.SpendToken(1)
		if err := client.DeleteStartupScript(string(d.scriptID)); err != nil {
			if strings.Contains(err.Error(), "Check SCRIPTID") {
				log.Infof("Script doesn't exist, assuming it is already deleted")
			} else {
				return err
			}
		}
	}
	return nil
}

func (d *Driver) Restart() error {
	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Stopped {
		log.Infof("Host is already stopped, use start command to run it")
		return nil
	}
	log.Debugf("restarting %s", d.MachineName)
	<-d.bucket.SpendToken(1)
	return d.getClient().RebootServer(d.MachineID)
}

func (d *Driver) Kill() error {
	if vmState, err := d.GetState(); err != nil {
		return err
	} else if vmState == state.Stopped {
		log.Infof("Host is already stopped")
		return nil
	}
	log.Debugf("killing %s", d.MachineName)
	<-d.bucket.SpendToken(1)
	return d.getClient().HaltServer(d.MachineID)
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

func (d *Driver) instanceIsRunning() bool {
	st, err := d.GetState()
	if err != nil {
		log.Debug(err)
	}
	if st == state.Running {
		return true
	}
	log.Debug("VPS not yet started")
	return false
}

func (d *Driver) validateApiCredentials() error {
	_, err := d.getClient().GetAccountInfo()
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) validateRegion() error {
	<-d.bucket.SpendToken(1)
	regions, err := d.getClient().GetRegions()
	if err != nil {
		return err
	}
	for _, region := range regions {
		if region.ID == d.RegionID {
			return nil
		}
	}
	return fmt.Errorf("Region ID %d is invalid", d.RegionID)
}

func (d *Driver) validatePlan() error {
	<-d.bucket.SpendToken(1)
	plans, err := d.getClient().GetAvailablePlansForRegion(d.RegionID)
	if err != nil {
		return err
	}
	for _, v := range plans {
		if v == d.PlanID {
			return nil
		}
	}
	return fmt.Errorf("Plan ID %d not available in this region. Available plans: %v", d.PlanID, plans)
}

// RancherOS - Create iPXE boot script
func (d *Driver) createBootScript() error {
	content := `#!ipxe
set base-url https://releases.rancher.com/os/latest
kernel ${base-url}/vmlinuz rancher.state.autoformat=[/dev/vda] rancher.cloud_init.datasources=[ec2]
initrd ${base-url}/initrd
boot`
	<-d.bucket.SpendToken(1)
	script, err := d.getClient().CreateStartupScript(d.MachineName, content, "pxe")
	if err != nil {
		return err
	}
	d.scriptID, err = strconv.Atoi(script.ID)
	if err != nil {
		return err
	}
	return nil
}

// RancherOS - Generate cloud-config userdata string that will
// provision the SSH Key to the VPS and configure private networking
func (d *Driver) getCloudConfig() (string, error) {
	type userData struct {
		HostName   string
		SSHkey     string
		PrivateNet bool
	}
	const tpl = `#cloud-config
hostname: {{.HostName}}
ssh_authorized_keys:
  - {{.SSHkey}}{{if .PrivateNet}}
write_files:
  - path: /opt/rancher/bin/start.sh
    permissions: 0700
    content: |
      #!/bin/bash
      sudo netconf
      rm -- "$0"
rancher:
  network:
    interfaces:
      eth0:
        dhcp: true
      eth1:
        address: $private_ipv4/16
        mtu: 1450{{end}}
`
	var buffer bytes.Buffer

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}
	config := userData{HostName: d.MachineName, SSHkey: string(publicKey), PrivateNet: d.PrivateNetworking}

	tmpl, err := template.New("cloud-config").Parse(tpl)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buffer, config)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
