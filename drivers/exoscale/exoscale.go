package exoscale

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/state"
	"github.com/pyr/egoscale/src/egoscale"
)

type Driver struct {
	URL              string
	ApiKey           string
	ApiSecretKey     string
	InstanceProfile  string
	DiskSize         int
	Image            string
	SecurityGroup    string
	AvailabilityZone string
	MachineName      string
	KeyPair          string
	IPAddress        string
	PublicKey        string
	Id               string
	CaCertPath       string
	PrivateKeyPath   string
	DriverKeyPath    string
	SwarmMaster      bool
	SwarmHost        string
	SwarmDiscovery   string
	storePath        string
}

func init() {
	drivers.Register("exoscale", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "EXOSCALE_ENDPOINT",
			Name:   "exoscale-url",
			Usage:  "exoscale API endpoint",
		},
		cli.StringFlag{
			EnvVar: "EXOSCALE_API_KEY",
			Name:   "exoscale-api-key",
			Usage:  "exoscale API key",
		},
		cli.StringFlag{
			EnvVar: "EXOSCALE_API_SECRET",
			Name:   "exoscale-api-secret-key",
			Usage:  "exoscale API secret key",
		},
		cli.StringFlag{
			EnvVar: "EXOSCALE_INSTANCE_PROFILE",
			Name:   "exoscale-instance-profile",
			Value:  "small",
			Usage:  "exoscale instance profile (small, medium, large, ...)",
		},
		cli.IntFlag{
			EnvVar: "EXOSCALE_DISK_SIZE",
			Name:   "exoscale-disk-size",
			Value:  50,
			Usage:  "exoscale disk size (10, 50, 100, 200, 400)",
		},
		cli.StringFlag{
			EnvVar: "EXSOCALE_IMAGE",
			Name:   "exoscale-image",
			Value:  "ubuntu-14.04",
			Usage:  "exoscale image template",
		},
		cli.StringFlag{
			EnvVar: "EXOSCALE_SECURITY_GROUP",
			Name:   "exoscale-security-group",
			Value:  "docker-machine",
			Usage:  "exoscale security group",
		},
		cli.StringFlag{
			EnvVar: "EXOSCALE_AVAILABILITY_ZONE",
			Name:   "exoscale-availability-zone",
			Value:  "ch-gva-2",
			Usage:  "exoscale availibility zone",
		},
		cli.StringFlag{
			EnvVar: "EXOSCALE_KEYPAIR",
			Name:   "exoscale-keypair",
			Usage:  "exoscale keypair name",
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
	return 22, nil
}

func (d *Driver) GetSSHUsername() string {
	return "ubuntu"
}

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Remote
}

func (d *Driver) DriverName() string {
	return "exoscale"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.URL = flags.String("exoscale-endpoint")
	d.ApiKey = flags.String("exoscale-api-key")
	d.ApiSecretKey = flags.String("exoscale-api-secret-key")
	d.InstanceProfile = flags.String("exoscale-instance-profile")
	d.DiskSize = flags.Int("exoscale-disk-size")
	d.Image = flags.String("exoscale-image")
	d.SecurityGroup = flags.String("exoscale-security-group")
	d.AvailabilityZone = flags.String("exoscale-availability-zone")
	d.KeyPair = flags.String("exoscale-keypair")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	if d.URL == "" {
		d.URL = "https://api.exoscale.ch/compute"
	}
	if d.ApiKey == "" || d.ApiSecretKey == "" {
		return fmt.Errorf("Please specify an API key (--exoscale-api-key) and an API secret key (--exoscale-api-secret-key).")
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
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	client := egoscale.NewClient(d.URL, d.ApiKey, d.ApiSecretKey)
	vm, err := client.GetVirtualMachine(d.Id)
	if err != nil {
		return state.Error, err
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

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	log.Infof("Querying exoscale for the requested parameters...")
	client := egoscale.NewClient(d.URL, d.ApiKey, d.ApiSecretKey)
	topology, err := client.GetTopology()
	if err != nil {
		return err
	}

	// Availability zone UUID
	zone, ok := topology.Zones[d.AvailabilityZone]
	if !ok {
		return fmt.Errorf("Availability zone %v doesn't exist",
			d.AvailabilityZone)
	}
	log.Debugf("Availability zone %v = %s", d.AvailabilityZone, zone)

	// Image UUID
	var tpl string
	images, ok := topology.Images[strings.ToLower(d.Image)]
	if ok {
		tpl, ok = images[d.DiskSize]
	}
	if !ok {
		return fmt.Errorf("Unable to find image %v with size %d",
			d.Image, d.DiskSize)
	}
	log.Debugf("Image %v(%d) = %s", d.Image, d.DiskSize, tpl)

	// Profile UUID
	profile, ok := topology.Profiles[strings.ToLower(d.InstanceProfile)]
	if !ok {
		return fmt.Errorf("Unable to find the %s profile",
			d.InstanceProfile)
	}
	log.Debugf("Profile %v = %s", d.InstanceProfile, profile)

	// Security group
	sg, ok := topology.SecurityGroups[d.SecurityGroup]
	if !ok {
		log.Infof("Security group %v does not exist, create it",
			d.SecurityGroup)
		rules := []egoscale.SecurityGroupRule{
			{
				SecurityGroupId: "",
				Cidr:            "0.0.0.0/0",
				Protocol:        "TCP",
				Port:            22,
			},
			{
				SecurityGroupId: "",
				Cidr:            "0.0.0.0/0",
				Protocol:        "TCP",
				Port:            2376,
			},
			{
				SecurityGroupId: "",
				Cidr:            "0.0.0.0/0",
				Protocol:        "ICMP",
				IcmpType:        8,
				IcmpCode:        0,
			},
		}
		sgresp, err := client.CreateSecurityGroupWithRules(d.SecurityGroup,
			rules,
			make([]egoscale.SecurityGroupRule, 0, 0))
		if err != nil {
			return err
		}
		sg = sgresp.Id
	}
	log.Debugf("Security group %v = %s", d.SecurityGroup, sg)

	if d.KeyPair == "" {
		log.Infof("Generate an SSH keypair...")
		kpresp, err := client.CreateKeypair(d.MachineName)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(d.GetSSHKeyPath(), []byte(kpresp.Privatekey), 0600)
		if err != nil {
			return err
		}
		d.KeyPair = d.MachineName
	}

	log.Infof("Spawn exoscale host...")

	userdata, err := d.getCloudInit()
	if err != nil {
		return err
	}
	log.Debugf("Using the following cloud-init file:")
	log.Debugf("%s", userdata)

	machineProfile := egoscale.MachineProfile{
		Template:        tpl,
		ServiceOffering: profile,
		SecurityGroups:  []string{sg},
		Userdata:        userdata,
		Zone:            zone,
		Keypair:         d.KeyPair,
		Name:            d.MachineName,
	}

	cvmresp, err := client.CreateVirtualMachine(machineProfile)
	if err != nil {
		return err
	}

	vm, err := d.waitForVM(client, cvmresp)
	if err != nil {
		return err
	}
	d.IPAddress = vm.Nic[0].Ipaddress
	d.Id = vm.Id

	return nil
}

func (d *Driver) Start() error {
	vmstate, err := d.GetState()
	if err != nil {
		return err
	}
	if vmstate == state.Running || vmstate == state.Starting {
		log.Infof("Host is already running or starting")
		return nil
	}

	client := egoscale.NewClient(d.URL, d.ApiKey, d.ApiSecretKey)
	svmresp, err := client.StartVirtualMachine(d.Id)
	if err != nil {
		return err
	}
	_, err = d.waitForVM(client, svmresp)
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
		log.Infof("Host is already stopped")
		return nil
	}

	client := egoscale.NewClient(d.URL, d.ApiKey, d.ApiSecretKey)
	svmresp, err := client.StopVirtualMachine(d.Id)
	if err != nil {
		return err
	}
	_, err = d.waitForVM(client, svmresp)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) Remove() error {
	client := egoscale.NewClient(d.URL, d.ApiKey, d.ApiSecretKey)
	dvmresp, err := client.DestroyVirtualMachine(d.Id)
	if err != nil {
		return err
	}
	_, err = d.waitForVM(client, dvmresp)
	if err != nil {
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
		return fmt.Errorf("Host is stopped, use start command to start it")
	}

	client := egoscale.NewClient(d.URL, d.ApiKey, d.ApiSecretKey)
	svmresp, err := client.RebootVirtualMachine(d.Id)
	if err != nil {
		return err
	}
	_, err = d.waitForVM(client, svmresp)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) waitForVM(client *egoscale.Client, jobid string) (*egoscale.DeployVirtualMachineResponse, error) {
	log.Infof("Waiting for VM...")
	maxRepeats := 60
	i := 0
	var resp *egoscale.QueryAsyncJobResultResponse
	var err error
	for ; i < maxRepeats; i++ {
		resp, err = client.PollAsyncJob(jobid)
		if err != nil {
			return nil, err
		}

		if resp.Jobstatus == 1 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if i == maxRepeats {
		return nil, fmt.Errorf("Timeout while waiting for VM")
	}
	vm, err := client.AsyncToVirtualMachine(*resp)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// Build a cloud-init user data string that will install and run
// docker.
func (d *Driver) getCloudInit() (string, error) {
	const tpl = `#cloud-config
manage_etc_hosts: true
fqdn: {{ .MachineName }}
resize_rootfs: true
`
	var buffer bytes.Buffer

	tmpl, err := template.New("cloud-init").Parse(tpl)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buffer, d)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
