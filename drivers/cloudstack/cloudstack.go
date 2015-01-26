package cloudstack

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

const (
	driverName      = "cloudstack"
	dockerConfigDir = "/etc/docker"
)

type configError struct {
	option string
}

func (e *configError) Error() string {
	return fmt.Sprintf("cloudstack driver requires the --cloudstack-%s option", e.option)
}

type unsupportedError struct {
	operation string
}

func (e *unsupportedError) Error() string {
	return fmt.Sprintf("CloudStack does not currently support the %q operation", e.operation)
}

type Driver struct {
	Id             string
	ApiURL         string
	ApiKey         string
	SecretKey      string
	MachineName    string
	NoPublicIP     bool
	PublicIP       string
	PublicPort     int
	PublicSSHPort  int
	PrivateIP      string
	PrivatePort    int
	SourceCIDR     string
	FWRuleIds      []string
	Explunge       bool
	Template       string
	Offering       string
	Network        string
	Zone           string
	CaCertPath     string
	PrivateKeyPath string

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
		cli.IntFlag{
			Name:  "cloudstack-public-ssh-port",
			Usage: "Public SSH port, if empty defaults to port 22",
		},
		cli.IntFlag{
			Name:  "cloudstack-private-port",
			Usage: "Private port, if empty defaults to port 2376",
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

func NewDriver(
	machineName string,
	storePath string,
	caCertPath string,
	privateKeyPath string) (drivers.Driver, error) {

	driver := &Driver{
		MachineName:    machineName,
		FWRuleIds: 			[]string{},
		storePath:      storePath,
		CaCertPath:     caCertPath,
		PrivateKeyPath: privateKeyPath,
	}
	return driver, nil
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ApiURL = flags.String("cloudstack-api-url")
	d.ApiKey = flags.String("cloudstack-api-key")
	d.SecretKey = flags.String("cloudstack-secret-key")
	d.NoPublicIP = flags.Bool("cloudstack-no-public-ip")
	d.PublicIP = flags.String("cloudstack-public-ip")
	d.PublicPort = flags.Int("cloudstack-public-port")
	d.PublicSSHPort = flags.Int("cloudstack-public-ssh-port")
	d.PrivatePort = flags.Int("cloudstack-private-port")
	d.SourceCIDR = flags.String("cloudstack-source-cidr")
	d.Explunge = flags.Bool("cloudstack-explunge")
	d.Template = flags.String("cloudstack-template")
	d.Offering = flags.String("cloudstack-offering")
	d.Network = flags.String("cloudstack-network")
	d.Zone = flags.String("cloudstack-zone")

	if d.ApiURL == "" {
		return &configError{option: "api-url"}
	}

	if d.ApiKey == "" {
		return &configError{option: "api-key"}
	}

	if d.SecretKey == "" {
		return &configError{option: "secret-key"}
	}

	if !d.NoPublicIP && d.PublicIP == "" {
		return &configError{option: "public-ip"}
	}

	if d.Template == "" {
		return &configError{option: "template"}
	}

	if d.Offering == "" {
		return &configError{option: "offering"}
	}

	if d.Zone == "" {
		return &configError{option: "zone"}
	}

	if d.PrivatePort == 0 {
		d.PrivatePort = 2376
	}

	if d.PublicPort == 0 {
		d.PublicPort = d.PrivatePort
	}

	if d.PublicSSHPort == 0 {
		d.PublicSSHPort = 22
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
	return fmt.Sprintf("tcp://%s:%d", ip, d.PublicPort), nil
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
	log.Info("Creating CloudStack instance...")

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

	log.Infof("Retrieving UUID of zone: %q", d.Zone)
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

		// Open the firewall for docker traffic
		r := cs.Firewall.NewCreateFirewallRuleParams(ipaddressid, "tcp")
		r.SetCidrlist([]string{d.SourceCIDR})
		r.SetStartport(d.PublicPort)
		r.SetEndport(d.PublicPort)

		fr, err := cs.Firewall.CreateFirewallRule(r)
		if err != nil {
			// Check if the error reports the port is already open
			if !strings.Contains(err.Error(), fmt.Sprintf(
				"The range specified, %d-%d, conflicts with rule", d.PublicPort, d.PublicPort)) {
				return err
			}
		} else {
			// Only set this if there really is anything to set :)
			d.FWRuleIds = append(d.FWRuleIds, fr.Id)
		}

		// And of course also for SSH traffic
		r.SetStartport(d.PublicSSHPort)
		r.SetEndport(d.PublicSSHPort)
		fr, err = cs.Firewall.CreateFirewallRule(r)
		if err != nil {
			// Check if the error reports the port is already open
			if !strings.Contains(err.Error(), fmt.Sprintf(
				"The range specified, %d-%d, conflicts with rule", d.PublicSSHPort, d.PublicSSHPort)) {
				return err
			}
		} else {
			// Only set this if there really is anything to set :)
			d.FWRuleIds = append(d.FWRuleIds, fr.Id)
		}

		// Create a port forward for the docker port
		f := cs.Firewall.NewCreatePortForwardingRuleParams(
			ipaddressid, d.PrivatePort, "tcp", d.PublicPort, d.Id)
		f.SetOpenfirewall(false)

		_, err = cs.Firewall.CreatePortForwardingRule(f)
		if err != nil {
			return err
		}

		// And another one for the SSH port
		f.SetPrivateport(22)
		f.SetPublicport(d.PublicSSHPort)
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
		log.Info("Machine is already running")
		return nil
	}

	if vmstate == state.Starting {
		log.Info("Machine is already starting")
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
		log.Info("Machine is already stopped")
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
	if len(d.FWRuleIds) > 0 {
		for _, id := range d.FWRuleIds {
			f := cs.Firewall.NewDeleteFirewallRuleParams(id)
			_, err = cs.Firewall.DeleteFirewallRule(f)
			if err != nil {
				return err
			}
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

func (d *Driver) StartDocker() error {
	log.Info("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo systemctl start docker.service")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) StopDocker() error {
	log.Info("Stopping Docker...")

	cmd, err := d.GetSSHCommand("sudo systemctl stop docker.service")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetDockerConfigDir() string {
	return dockerConfigDir
}

func (d *Driver) Upgrade() error {
	return &unsupportedError{operation: "upgrade"}
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ipaddress, err := d.GetIP()
	if err != nil {
		return nil, err
	}
	return ssh.GetSSHCommand(ipaddress, d.PublicSSHPort, "core", d.sshKeyPath(), args...), nil
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
      Environment="DOCKER_OPTS='--auth=identity --host=tcp://0.0.0.0:%d'"
  - path: /.docker/authorized-keys.d/docker-host.json
    owner: core:core
    permissions: '0644'
    content: |
      %s

ssh-authorized-keys:
  - %s

coreos:
  units:
    - name: docker-tcp.socket
      command: start
      enable: yes
      content: |
        [Unit]
        Description=Docker Socket for the API

        [Socket]
        ListenStream=%d
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
`, d.PrivatePort, pubKey, pubKey, d.PrivatePort)

	ud := base64.StdEncoding.EncodeToString([]byte(cc))
	if len(ud) > 32768 {
		return "", fmt.Errorf(
			"The created user_data contains %d bytes after encoding, "+
				"this exeeds the limit of 32768 bytes", len(ud))
	}

	return ud, nil
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}
