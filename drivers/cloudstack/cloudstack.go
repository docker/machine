package cloudstack

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"os/exec"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

const (
	driverName = "cloudstack"
)

type Driver struct {
	Id          string
	ApiURL      string
	ApiKey      string
	SecretKey   string
	MachineName string
	NoPublicIP  bool
	PublicIP    string
	PublicPort  int
	PrivateIP   string
	SourceCIDR  string
	FWRuleId    string
	Explunge    bool
	Template    string
	Offering    string
	Network     string
	Zone        string

	storePath string
}

func init() {
	drivers.Register(driverName, &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "CLOUDSTACK_API_URL",
			Name:   "cloudstack-api-url",
			Usage:  "CloudStack API URL",
		},
		cli.StringFlag{
			EnvVar: "CLOUDSTACK_API_KEY",
			Name:   "cloudstack-api-key",
			Usage:  "CloudStack API key",
		},
		cli.StringFlag{
			EnvVar: "CLOUDSTACK_SECRET_KEY",
			Name:   "cloudstack-secret-key",
			Usage:  "CloudStack API secret key",
		},
		cli.StringFlag{
			Name:  "cloudstack-machinename",
			Usage: "Machine name",
		},
		cli.BoolFlag{
			Name: "cloudstack-no-public-ip",
			Usage: "Do not use a public IP for this host, helpfull in cases where you have direct " +
				"access to the IP addresses assigned by DHCP",
		},
		cli.StringFlag{
			Name:  "cloudstack-public-ip",
			Usage: "Public IP", //leave empty to assign a new public IP to the machine",
		},
		cli.IntFlag{
			Name:  "cloudstack-public-port",
			Usage: "Public port, if empty matches the private port",
		},
		cli.StringFlag{
			Name:  "cloudstack-source-cidr",
			Usage: "CIDR block to give access to the new machine, leave empty to use 0.0.0.0/0",
		},
		cli.BoolFlag{
			Name:  "cloudstack-explunge",
			Usage: "Whether or not to explunge the machine upon removal",
		},
		cli.StringFlag{
			Name:  "cloudstack-template",
			Usage: "CloudStack template",
		},
		cli.StringFlag{
			Name:  "cloudstack-offering",
			Usage: "CloudStack service offering",
		},
		cli.StringFlag{
			Name:  "cloudstack-network",
			Usage: "CloudStack network",
		},
		cli.StringFlag{
			Name:  "cloudstack-zone",
			Usage: "CloudStack zone",
		},
	}
}

func NewDriver(storePath string) (drivers.Driver, error) {
	return &Driver{storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ApiURL = flags.String("cloudstack-api-url")
	d.ApiKey = flags.String("cloudstack-api-key")
	d.SecretKey = flags.String("cloudstack-secret-key")
	d.MachineName = flags.String("cloudstack-machinename")
	d.NoPublicIP = flags.Bool("cloudstack-no-public-ip")
	d.PublicIP = flags.String("cloudstack-public-ip")
	d.PublicPort = flags.Int("cloudstack-public-port")
	d.SourceCIDR = flags.String("cloudstack-source-cidr")
	d.Explunge = flags.Bool("cloudstack-explunge")
	d.Template = flags.String("cloudstack-template")
	d.Offering = flags.String("cloudstack-offering")
	d.Network = flags.String("cloudstack-network")
	d.Zone = flags.String("cloudstack-zone")

	if d.ApiURL == "" {
		return fmt.Errorf("cloudstack driver requires the --cloudstack-api-url option")
	}

	if d.ApiKey == "" {
		return fmt.Errorf("cloudstack driver requires the --cloudstack-api-key option")
	}

	if d.SecretKey == "" {
		return fmt.Errorf("cloudstack driver requires the --cloudstack-secret-key option")
	}

	if !d.NoPublicIP && d.PublicIP == "" {
		return fmt.Errorf("cloudstack driver requires the --cloudstack-public-ip option")
	}

	if d.Template == "" {
		return fmt.Errorf("cloudstack driver requires the --cloudstack-template option")
	}

	if d.Offering == "" {
		return fmt.Errorf("cloudstack driver requires the --cloudstack-offering option")
	}

	if d.Zone == "" {
		return fmt.Errorf("cloudstack driver requires the --cloudstack-zone option")
	}

	if d.MachineName == "" {
		rand.Seed(time.Now().UnixNano())
		d.MachineName = fmt.Sprintf("docker-host-%04x", rand.Intn(65535))
	}

	if d.PublicPort == 0 {
		d.PublicPort = 2376
	}

	if d.SourceCIDR == "" {
		d.SourceCIDR = "0.0.0.0/0"
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.NoPublicIP {
		return d.PrivateIP, nil
	}
	return d.PublicIP, nil
}

func (d *Driver) GetState() (state.State, error) {
	cs := d.getClient()
	vm, count, err := cs.VirtualMachine.GetVirtualMachineByID(d.Id)
	if err != nil {
		return state.Error, err
	}

	if count == 0 {
		return state.None, fmt.Errorf("Machine doesn not exist, use create command to create it")
	}

	switch vm.State {
	case "Starting":
		return state.Starting, nil
	case "Running":
		return state.Running, nil
	case "Stopping":
		return state.Running, nil
	case "Stopped":
		return state.Stopped, nil
	case "Destroyed":
		return state.Stopped, nil
	case "Expunging":
		return state.Stopped, nil
	case "Migrating":
		return state.Paused, nil
	case "Error":
		return state.Error, nil
	case "Unknown":
		return state.Error, nil
	case "Shutdowned":
		return state.Stopped, nil
	}

	return state.None, nil
}

func (d *Driver) Create() error {
	log.Infof("Creating CloudStack instance...")

	cs := d.getClient()

	log.Infof("Retrieving UUID of template: %q", d.Template)
	templateid, err := cs.Template.GetTemplateID(d.Template, "executable")
	if err != nil {
		return fmt.Errorf("Error retrieving UUID of template %q: %v", d.Template, err)
	}

	log.Infof("Retrieving UUID of service offering: %q", d.Offering)
	offeringid, err := cs.ServiceOffering.GetServiceOfferingID(d.Offering)
	if err != nil {
		return fmt.Errorf("Error retrieving UUID of service offering %q: %v", d.Offering, err)
	}

	log.Infof("Retrieving UUID of zone: %q", d.Template)
	zoneid, err := cs.Zone.GetZoneID(d.Zone)
	if err != nil {
		return fmt.Errorf("Error retrieving UUID of zone %q: %v", d.Zone, err)
	}

	p := cs.VirtualMachine.NewDeployVirtualMachineParams(offeringid, templateid, zoneid)
	p.SetName(d.MachineName)
	p.SetDisplayname(d.MachineName)

	if d.Network != "" {
		log.Infof("Retrieving UUID of network: %q", d.Network)
		networkid, err := cs.Network.GetNetworkID(d.Network)
		if err != nil {
			return fmt.Errorf("Error retrieving UUID of network %q: %v", d.Network, err)
		}
		p.SetNetworkids([]string{networkid})
	}

	log.Infof("Creating SSH key pair...")
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	ud, err := d.getUserData()
	if err != nil {
		return err
	}
	p.SetUserdata(ud)

	// If no IP address is given, make sure we assign one
	if !d.NoPublicIP && d.PublicIP == "" {
		// This needs to be implemented, for now we assume a machine is always
		// created in a network that already has a public IP and so we only need
		// to open and forward the correct ports in the firewall...
	}

	// Create the machine
	vm, err := cs.VirtualMachine.DeployVirtualMachine(p)
	if err != nil {
		return err
	}

	d.Id = vm.Id
	d.PrivateIP = vm.Nic[0].Ipaddress

	if !d.NoPublicIP {
		// Make sure the new machine is accessible
		log.Infof("Retrieving UUID of IP address: %q", d.PublicIP)
		ip := cs.Address.NewListPublicIpAddressesParams()
		ip.SetIpaddress(d.PublicIP)

		l, err := cs.Address.ListPublicIpAddresses(ip)
		if err != nil {
			return err
		}
		if l.Count != 1 {
			return fmt.Errorf("Could not find UUID of IP address: %s", d.PublicIP)
		}
		ipaddressid := l.PublicIpAddresses[0].Id

		r := cs.Firewall.NewCreateFirewallRuleParams(ipaddressid, "tcp")
		r.SetCidrlist([]string{d.SourceCIDR})
		r.SetStartport(d.PublicPort)
		r.SetEndport(d.PublicPort)

		rr, err := cs.Firewall.CreateFirewallRule(r)
		if err != nil {
			// Check if the error reports the port is already open
			if !strings.Contains(err.Error(), fmt.Sprintf(
				"The range specified, %d-%d, conflicts with rule", d.PublicPort, d.PublicPort)) {
				return err
			}
		} else {
			// Only set this if there really is anything to set :)
			d.FWRuleId = rr.Id
		}

		f := cs.Firewall.NewCreatePortForwardingRuleParams(
			ipaddressid, 2376, "tcp", d.PublicPort, d.Id)
		f.SetOpenfirewall(false)

		_, err = cs.Firewall.CreatePortForwardingRule(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) Start() error {
	vmstate, err := d.GetState()
	if err != nil {
		return err
	}

	if vmstate == state.Running {
		log.Infof("Machine is already running")
		return nil
	}

	if vmstate == state.Starting {
		log.Infof("Machine is already starting")
		return nil
	}

	cs := d.getClient()
	p := cs.VirtualMachine.NewStartVirtualMachineParams(d.Id)

	_, err = cs.VirtualMachine.StartVirtualMachine(p)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	vmstate, err := d.GetState()
	if err != nil {
		return err
	}

	if vmstate == state.Stopped {
		log.Infof("Machine is already stopped")
		return nil
	}

	cs := d.getClient()
	p := cs.VirtualMachine.NewStopVirtualMachineParams(d.Id)

	_, err = cs.VirtualMachine.StopVirtualMachine(p)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Remove() error {
	cs := d.getClient()
	p := cs.VirtualMachine.NewDestroyVirtualMachineParams(d.Id)
	p.SetExpunge(d.Explunge)

	_, err := cs.VirtualMachine.DestroyVirtualMachine(p)
	if err != nil {
		return err
	}

	// Not sure if this should be here, I can imagine some use cases
	// where does shouldn't be executed
	if d.FWRuleId != "" {
		f := cs.Firewall.NewDeleteFirewallRuleParams(d.FWRuleId)
		_, err = cs.Firewall.DeleteFirewallRule(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) Restart() error {
	vmstate, err := d.GetState()
	if err != nil {
		return err
	}

	if vmstate == state.Stopped {
		return fmt.Errorf("Machine is stopped, use start command to start it")
	}

	cs := d.getClient()
	p := cs.VirtualMachine.NewRebootVirtualMachineParams(d.Id)

	_, err = cs.VirtualMachine.RebootVirtualMachine(p)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("hosts without a driver cannot be upgraded")
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return nil, fmt.Errorf("hosts without a driver do not support SSH")
}

func (d *Driver) getClient() *cloudstack.CloudStackClient {
	cs := cloudstack.NewAsyncClient(d.ApiURL, d.ApiKey, d.SecretKey, false)
	cs.AsyncTimeout(180)
	return cs
}

func (d *Driver) getUserData() (string, error) {
	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", fmt.Errorf("Error reading public key: %v", err)
	}

	cc := fmt.Sprintf(`#cloud-config
write_files:
  - path: /etc/systemd/system/docker.service.d/auth-identity.conf
    owner: core:core
    permissions: 0644
    content: |
      [Service]
      Environment="DOCKER_OPTS='--auth=identity --host=tcp://0.0.0.0:2376'"
  - path: /.docker/authorized-keys.d/docker-host.json
    owner: root:root
    permissions: '0644'
    content: |
      %s

coreos:
  units:
    - name: docker-tcp.socket
      command: start
      enable: yes
      content: |
        [Unit]
        Description=Docker Socket for the API

        [Socket]
        ListenStream=2376
        BindIPv6Only=both
        Service=docker.service

        [Install]
        WantedBy=sockets.target
    - name: enable-docker-tcp.service
      command: start
      content: |
        [Unit]
        Description=Enable the Docker Socket for the API

        [Service]
        Type=oneshot
        ExecStart=/usr/bin/systemctl enable docker-tcp.socket
    - name: docker.service
      command: start
`, pubKey)

	ud := base64.StdEncoding.EncodeToString([]byte(cc))
	if len(ud) > 2048 {
		return "", fmt.Errorf(
			"The created user_data contains %d bytes after encoding, "+
				"this exeeds the limit of 2048 bytes", len(ud))
	}

	return ud, nil
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}
