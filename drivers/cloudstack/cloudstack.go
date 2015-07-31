package cloudstack

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/state"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

const (
	driverName = "cloudstack"
)

var (
	sshUser    = "ubuntu"
	sshPort    = 22
	dockerPort = 2376
	swarmPort  = 3376
)

type configError struct {
	option string
}

func (e *configError) Error() string {
	return fmt.Sprintf("cloudstack driver requires the --cloudstack-%s option", e.option)
}

type Driver struct {
	ID                string
	MachineName       string
	APIURL            string
	APIKey            string
	SecretKey         string
	Zone              string
	Template          string
	ServiceOffering   string
	Network           string
	PublicIP          string
	PublicIPID        string
	IPAddress         string
	SSHPort           int
	SSHUser           string
	SSHKeyPair        string
	SourceCIDR        string
	FWRuleIDs         []string
	PFRuleIDs         []string
	SecurityGroupName string
	Expunge           bool
	CaCertPath        string
	PrivateKeyPath    string
	SwarmMaster       bool
	SwarmHost         string
	SwarmDiscovery    string
	storePath         string
}

func init() {
	drivers.Register(driverName, &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func (d *Driver) DriverName() string {
	return driverName
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "cloudstack-api-url",
			Usage:  "CloudStack API URL",
			EnvVar: "CLOUDSTACK_API_URL",
		},
		cli.StringFlag{
			Name:   "cloudstack-api-key",
			Usage:  "CloudStack API key",
			EnvVar: "CLOUDSTACK_API_KEY",
		},
		cli.StringFlag{
			Name:   "cloudstack-secret-key",
			Usage:  "CloudStack API secret key",
			EnvVar: "CLOUDSTACK_SECRET_KEY",
		},
		cli.StringFlag{
			Name:  "cloudstack-zone",
			Usage: "CloudStack zone",
		},
		cli.StringFlag{
			Name:  "cloudstack-template",
			Usage: "CloudStack template",
		},
		cli.StringFlag{
			Name:  "cloudstack-service-offering",
			Usage: "CloudStack service offering",
		},
		cli.StringFlag{
			Name:  "cloudstack-network",
			Usage: "CloudStack network",
		},
		cli.StringFlag{
			Name:  "cloudstack-public-ip",
			Usage: "Public IP, leave empty to use Private IP",
		},
		cli.IntFlag{
			Name:  "cloudstack-ssh-port",
			Usage: "Public SSH port, if empty defaults to port 22",
			Value: sshPort,
		},
		cli.StringFlag{
			Name:  "cloudstack-ssh-user",
			Usage: "CloudStack SSH User",
			Value: sshUser,
		},
		cli.StringFlag{
			Name:  "cloudstack-source-cidr",
			Usage: "CIDR block to give access to the new machine",
			Value: "0.0.0.0/0",
		},
		cli.BoolFlag{
			Name:  "cloudstack-expunge",
			Usage: "Whether or not to expunge the machine upon removal",
		},
	}
}

func NewDriver(
	machineName string,
	storePath string,
	caCert string,
	privateKey string) (drivers.Driver, error) {
	return &Driver{
		MachineName:    machineName,
		storePath:      storePath,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
	}, nil
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	// Required
	d.APIURL = flags.String("cloudstack-api-url")
	d.APIKey = flags.String("cloudstack-api-key")
	d.SecretKey = flags.String("cloudstack-secret-key")
	d.ServiceOffering = flags.String("cloudstack-service-offering")
	d.Template = flags.String("cloudstack-template")
	d.Zone = flags.String("cloudstack-zone")

	// Optional
	d.Network = flags.String("cloudstack-network")
	d.PublicIP = flags.String("cloudstack-public-ip")
	d.SSHPort = flags.Int("cloudstack-ssh-port")
	d.SSHUser = flags.String("cloudstack-ssh-user")
	d.SourceCIDR = flags.String("cloudstack-source-cidr")
	d.Expunge = flags.Bool("cloudstack-expunge")

	// Swarm
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	if d.APIURL == "" {
		return &configError{option: "api-url"}
	}

	if d.APIKey == "" {
		return &configError{option: "api-key"}
	}

	if d.SecretKey == "" {
		return &configError{option: "secret-key"}
	}

	if d.Template == "" {
		return &configError{option: "template"}
	}

	if d.ServiceOffering == "" {
		return &configError{option: "service-offering"}
	}

	if d.Zone == "" {
		return &configError{option: "zone"}
	}

	ipaddressid, err := d.getPublicIPID()
	if err == nil {
		d.PublicIPID = ipaddressid
	}

	if d.SwarmMaster {
		u, err := url.Parse(d.SwarmHost)
		if err != nil {
			return fmt.Errorf("error parsing swarm host: %s", err)
		}

		parts := strings.Split(u.Host, ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		swarmPort = port
	}

	return nil
}

func (d *Driver) getPublicIPID() (string, error) {
	if d.PublicIP == "" {
		return "", nil
	}

	cs := d.getClient()

	log.Debugf("Retrieving UUID of IP address: %q", d.PublicIP)
	p := cs.Address.NewListPublicIpAddressesParams()
	p.SetIpaddress(d.PublicIP)

	l, err := cs.Address.ListPublicIpAddresses(p)
	if err != nil {
		return "", err
	}
	if l.Count != 1 {
		return "", fmt.Errorf("Could not find UUID of IP address: %s", d.PublicIP)
	}

	return l.PublicIpAddresses[0].Id, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil || ip == "" {
		return "", err
	}

	return fmt.Sprintf("tcp://%s:%d", ip, dockerPort), nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetIP() (string, error) {
	if d.PublicIP == "" {
		return d.IPAddress, nil
	}
	return d.PublicIP, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHPort() (int, error) {
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetState() (state.State, error) {
	cs := d.getClient()
	vm, count, err := cs.VirtualMachine.GetVirtualMachineByID(d.ID)
	if err != nil {
		return state.Error, err
	}

	if count == 0 {
		return state.None, fmt.Errorf("Machine does not exist, use create command to create it")
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

func (d *Driver) checkVMConflict() error {
	cs := d.getClient()

	p := cs.VirtualMachine.NewListVirtualMachinesParams()
	p.SetName(d.MachineName)

	r, err := cs.VirtualMachine.ListVirtualMachines(p)
	if err != nil {
		return err
	}
	if r.Count > 0 {
		return fmt.Errorf("Error machine %s already exists on CloudStack", d.MachineName)
	}

	return nil
}

func (d *Driver) checkSSHKeyPairConflict() error {
	cs := d.getClient()

	param := cs.SSH.NewListSSHKeyPairsParams()
	param.SetName(fmt.Sprintf("docker-machine-%s", d.MachineName))
	keypairs, err := cs.SSH.ListSSHKeyPairs(param)
	if err != nil {
		return err
	}
	if keypairs.Count > 0 {
		return fmt.Errorf("Error keypair %s already exists on CloudStack", d.MachineName)
	}

	return nil
}

func (d *Driver) checkFWPFRuleConflict() error {
	if d.PublicIPID == "" {
		return nil
	}

	cs := d.getClient()

	fwparam := cs.Firewall.NewListFirewallRulesParams()
	fwparam.SetIpaddressid(d.PublicIPID)
	fwrules, err := cs.Firewall.ListFirewallRules(fwparam)
	if err != nil {
		return err
	}

	pfparam := cs.Firewall.NewListPortForwardingRulesParams()
	pfparam.SetIpaddressid(d.PublicIPID)
	pfrules, err := cs.Firewall.ListPortForwardingRules(pfparam)
	if err != nil {
		return err
	}

	ports := []int{dockerPort, d.SSHPort}

	for _, port := range ports {
		for _, fw := range fwrules.FirewallRules {
			startPort, _ := strconv.Atoi(fw.Startport)
			endPort, _ := strconv.Atoi(fw.Endport)
			if fw.Protocol == "tcp" && startPort <= port && port <= endPort {
				return fmt.Errorf("Error port %d is already open by FWRule %s", port, fw.Id)
			}
		}

		for _, pf := range pfrules.PortForwardingRules {
			publicPort, _ := strconv.Atoi(pf.Publicport)
			if pf.Protocol == "tcp" && publicPort == port {
				return fmt.Errorf("Error port %d is already used by PFRule %s", port, pf.Id)
			}
		}
	}

	return nil
}

func (d *Driver) checkSecurityGroupConflict() error {
	cs := d.getClient()

	p := cs.SecurityGroup.NewListSecurityGroupsParams()
	p.SetSecuritygroupname(fmt.Sprintf("docker-machine-%s", d.MachineName))
	sgs, err := cs.SecurityGroup.ListSecurityGroups(p)
	if err != nil {
		return err
	}
	if sgs.Count > 0 {
		return fmt.Errorf("Error security group %s already exists on CloudStack", d.MachineName)
	}

	return nil
}

func (d *Driver) checkPrereqs() error {
	var err error

	if err = d.checkVMConflict(); err != nil {
		return err
	}

	if err = d.checkSSHKeyPairConflict(); err != nil {
		return err
	}

	if err = d.checkFWPFRuleConflict(); err != nil {
		return err
	}

	if err = d.checkSecurityGroupConflict(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	return d.checkPrereqs()
}

func (d *Driver) getNetworkType() (string, error) {
	cs := d.getClient()

	log.Debugf("Retrieving type of zone: %q", d.Zone)
	zone, _, err := cs.Zone.GetZoneByName(d.Zone)
	if err != nil {
		return "", fmt.Errorf("Error retrieving zone %q: %v", d.Zone, err)
	}

	return zone.Networktype, nil
}

func (d *Driver) createSSHKeyPair() error {
	cs := d.getClient()

	log.Info("Creating an SSH keypair...")
	keypairName := fmt.Sprintf("docker-machine-%s", d.MachineName)
	p := cs.SSH.NewCreateSSHKeyPairParams(keypairName)
	keypair, err := cs.SSH.CreateSSHKeyPair(p)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(d.GetSSHKeyPath(), []byte(keypair.Privatekey), 0600)
	if err != nil {
		return err
	}
	d.SSHKeyPair = keypairName

	return nil
}

func (d *Driver) deployVirtualMachine() error {
	cs := d.getClient()

	log.Debugf("Retrieving UUID of zone: %q", d.Zone)
	zoneid, err := cs.Zone.GetZoneID(d.Zone)
	if err != nil {
		return fmt.Errorf("Error retrieving UUID of zone %q: %v", d.Zone, err)
	}

	log.Debugf("Retrieving UUID of template: %q", d.Template)
	templateid, err := cs.Template.GetTemplateID(d.Template, "executable", zoneid)
	if err != nil {
		return fmt.Errorf("Error retrieving UUID of template %q: %v", d.Template, err)
	}

	log.Debugf("Retrieving UUID of service offering: %q", d.ServiceOffering)
	offeringid, err := cs.ServiceOffering.GetServiceOfferingID(d.ServiceOffering)
	if err != nil {
		return fmt.Errorf("Error retrieving UUID of service offering %q: %v", d.ServiceOffering, err)
	}

	networkid := ""
	if d.Network != "" {
		log.Debugf("Retrieving UUID of network: %q", d.Network)
		networkid, err = cs.Network.GetNetworkID(d.Network)
		if err != nil {
			return fmt.Errorf("Error retrieving UUID of network %q: %v", d.Network, err)
		}
	}

	p := cs.VirtualMachine.NewDeployVirtualMachineParams(offeringid, templateid, zoneid)
	p.SetName(d.MachineName)
	p.SetDisplayname(d.MachineName)
	p.SetKeypair(d.SSHKeyPair)

	if networkid != "" {
		p.SetNetworkids([]string{networkid})
	}

	if d.SecurityGroupName != "" {
		p.SetSecuritygroupnames([]string{d.SecurityGroupName})
	}

	log.Info("Creating CloudStack instance...")
	vm, err := cs.VirtualMachine.DeployVirtualMachine(p)
	if err != nil {
		return err
	}

	d.ID = vm.Id
	d.IPAddress = vm.Nic[0].Ipaddress

	return nil
}

func (d *Driver) createSecurityGroup() error {
	cs := d.getClient()

	log.Info("Creating security group...")
	sgName := fmt.Sprintf("docker-machine-%s", d.MachineName)
	p := cs.SecurityGroup.NewCreateSecurityGroupParams(sgName)
	if _, err := cs.SecurityGroup.CreateSecurityGroup(p); err != nil {
		return err
	}

	d.SecurityGroupName = sgName

	ports := []int{dockerPort, d.SSHPort}
	if d.SwarmMaster {
		ports = append(ports, swarmPort)
	}

	for _, port := range ports {
		p := cs.SecurityGroup.NewAuthorizeSecurityGroupIngressParams()
		p.SetSecuritygroupname(d.SecurityGroupName)
		p.SetCidrlist([]string{d.SourceCIDR})
		p.SetProtocol("tcp")
		p.SetStartport(port)
		p.SetEndport(port)
		if _, err := cs.SecurityGroup.AuthorizeSecurityGroupIngress(p); err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) createFirewallRules() error {
	cs := d.getClient()

	ports := []int{dockerPort, d.SSHPort}
	if d.SwarmMaster {
		ports = append(ports, swarmPort)
	}

	log.Info("Creating firewall rules...")
	for _, port := range ports {
		p := cs.Firewall.NewCreateFirewallRuleParams(d.PublicIPID, "tcp")
		p.SetCidrlist([]string{d.SourceCIDR})
		p.SetStartport(port)
		p.SetEndport(port)

		rule, err := cs.Firewall.CreateFirewallRule(p)
		if err != nil {
			return err
		}
		d.FWRuleIDs = append(d.FWRuleIDs, rule.Id)
	}

	return nil
}

func (d *Driver) createPortForwardingRules() error {
	cs := d.getClient()

	type pair struct {
		public, private int
	}

	ports := []pair{
		{public: dockerPort, private: dockerPort},
		{public: d.SSHPort, private: sshPort},
	}

	if d.SwarmMaster {
		ports = append(ports, pair{public: swarmPort, private: swarmPort})
	}

	log.Info("Creating port forwarding rules...")
	for _, port := range ports {
		p := cs.Firewall.NewCreatePortForwardingRuleParams(
			d.PublicIPID, port.private, "tcp", port.public, d.ID)
		p.SetOpenfirewall(false)
		rule, err := cs.Firewall.CreatePortForwardingRule(p)
		if err != nil {
			return err
		}

		d.PFRuleIDs = append(d.PFRuleIDs, rule.Id)
	}

	return nil
}

func (d *Driver) Create() error {
	if err := d.checkPrereqs(); err != nil {
		return err
	}

	nwType, err := d.getNetworkType()
	if err != nil {
		return err
	}

	if err := d.createSSHKeyPair(); err != nil {
		return err
	}

	if nwType == "Basic" {
		if err := d.createSecurityGroup(); err != nil {
			return err
		}
	}

	if err := d.deployVirtualMachine(); err != nil {
		return err
	}

	if nwType == "Advanced" && d.PublicIPID != "" {
		if err := d.createFirewallRules(); err != nil {
			return err
		}

		if err := d.createPortForwardingRules(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) Remove() error {
	cs := d.getClient()

	if d.SSHKeyPair != "" {
		log.Info("Removing ssh keypair...")
		p := cs.SSH.NewDeleteSSHKeyPairParams(d.SSHKeyPair)
		if _, err := cs.SSH.DeleteSSHKeyPair(p); err != nil {
			return err
		}
	}

	if len(d.PFRuleIDs) > 0 {
		log.Info("Removing port forwarding rules...")
		for _, id := range d.PFRuleIDs {
			p := cs.Firewall.NewDeletePortForwardingRuleParams(id)
			if _, err := cs.Firewall.DeletePortForwardingRule(p); err != nil {
				return err
			}
		}
	}

	if len(d.FWRuleIDs) > 0 {
		log.Info("Removing firewall rules...")
		for _, id := range d.FWRuleIDs {
			p := cs.Firewall.NewDeleteFirewallRuleParams(id)
			if _, err := cs.Firewall.DeleteFirewallRule(p); err != nil {
				return err
			}
		}
	}

	if d.ID != "" {
		log.Info("Removing CloudStack instance...")
		p := cs.VirtualMachine.NewDestroyVirtualMachineParams(d.ID)
		p.SetExpunge(d.Expunge)
		if _, err := cs.VirtualMachine.DestroyVirtualMachine(p); err != nil {
			return err
		}
	}

	if d.SecurityGroupName != "" && d.Expunge {
		log.Info("Removing CloudStack security group...")
		p := cs.SecurityGroup.NewDeleteSecurityGroupParams()
		p.SetName(d.SecurityGroupName)
		if _, err := cs.SecurityGroup.DeleteSecurityGroup(p); err != nil {
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
	p := cs.VirtualMachine.NewStartVirtualMachineParams(d.ID)
	if _, err = cs.VirtualMachine.StartVirtualMachine(p); err != nil {
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
	p := cs.VirtualMachine.NewStopVirtualMachineParams(d.ID)
	if _, err = cs.VirtualMachine.StopVirtualMachine(p); err != nil {
		return err
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
	p := cs.VirtualMachine.NewRebootVirtualMachineParams(d.ID)
	if _, err = cs.VirtualMachine.RebootVirtualMachine(p); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) getClient() *cloudstack.CloudStackClient {
	cs := cloudstack.NewAsyncClient(d.APIURL, d.APIKey, d.SecretKey, false)
	cs.AsyncTimeout(180)
	return cs
}
