package exoscale

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/exoscale/egoscale"
)

// Driver is the struct compatible with github.com/docker/machine/libmachine/drivers.Driver interface
type Driver struct {
	*drivers.BaseDriver
	URL              string
	APIKey           string `json:"ApiKey"`
	APISecretKey     string `json:"ApiSecretKey"`
	InstanceProfile  string
	DiskSize         int64
	Image            string
	SecurityGroup    string
	AffinityGroup    string
	AvailabilityZone string
	KeyPair          string
	PublicKey        string
	UserDataFile     string
	ID               string `json:"Id"`
	async            egoscale.AsyncInfo
}

const (
	defaultAPIEndpoint       = "https://api.exoscale.ch/compute"
	defaultInstanceProfile   = "Small"
	defaultDiskSize          = 50
	defaultImage             = "Linux Ubuntu 16.04 LTS 64-bit"
	defaultAvailabilityZone  = "CH-DK-2"
	defaultSSHUser           = "root"
	defaultAffinityGroupType = "host anti-affinity"
	defaultCloudInit         = `#cloud-config
manage_etc_hosts: localhost
`
)

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_ENDPOINT",
			Name:   "exoscale-url",
			Usage:  "exoscale API endpoint",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_API_KEY",
			Name:   "exoscale-api-key",
			Usage:  "exoscale API key",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_API_SECRET",
			Name:   "exoscale-api-secret-key",
			Usage:  "exoscale API secret key",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_INSTANCE_PROFILE",
			Name:   "exoscale-instance-profile",
			Value:  defaultInstanceProfile,
			Usage:  "exoscale instance profile (Small, Medium, Large, ...)",
		},
		mcnflag.IntFlag{
			EnvVar: "EXOSCALE_DISK_SIZE",
			Name:   "exoscale-disk-size",
			Value:  defaultDiskSize,
			Usage:  "exoscale disk size (10, 50, 100, 200, 400)",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_IMAGE",
			Name:   "exoscale-image",
			Value:  defaultImage,
			Usage:  "exoscale image template",
		},
		mcnflag.StringSliceFlag{
			EnvVar: "EXOSCALE_SECURITY_GROUP",
			Name:   "exoscale-security-group",
			Value:  []string{},
			Usage:  "exoscale security group",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_AVAILABILITY_ZONE",
			Name:   "exoscale-availability-zone",
			Value:  defaultAvailabilityZone,
			Usage:  "exoscale availability zone",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_SSH_USER",
			Name:   "exoscale-ssh-user",
			Value:  "",
			Usage:  "name of the ssh user",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_USERDATA",
			Name:   "exoscale-userdata",
			Usage:  "path to file with cloud-init user-data",
		},
		mcnflag.StringSliceFlag{
			EnvVar: "EXOSCALE_AFFINITY_GROUP",
			Name:   "exoscale-affinity-group",
			Value:  []string{},
			Usage:  "exoscale affinity group",
		},
	}
}

// NewDriver creates a Driver with the specified machineName and storePath.
func NewDriver(machineName, storePath string) drivers.Driver {
	return &Driver{
		InstanceProfile:  defaultInstanceProfile,
		DiskSize:         defaultDiskSize,
		Image:            defaultImage,
		AvailabilityZone: defaultAvailabilityZone,
		async: egoscale.AsyncInfo{
			Retries: 3,
			Delay:   20,
		},
		BaseDriver: &drivers.BaseDriver{
			MachineName: machineName,
			StorePath:   storePath,
		},
	}
}

// GetSSHHostname returns the hostname to use with SSH
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetSSHUsername returns the username to use with SSH
func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		name := strings.ToLower(d.Image)
		re := regexp.MustCompile(`\b[0-9.]+\b`)
		version := re.FindString(d.Image)

		if strings.Contains(name, "ubuntu") {
			return "ubuntu"
		}
		if strings.Contains(name, "centos") && version >= "7.3" {
			return "centos"
		}
		if strings.Contains(name, "coreos") {
			return "core"
		}
		if strings.Contains(name, "debian") && version >= "8" {
			return "debian"
		}
		return defaultSSHUser
	}

	return d.SSHUser
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "exoscale"
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.URL = flags.String("exoscale-url")
	d.APIKey = flags.String("exoscale-api-key")
	d.APISecretKey = flags.String("exoscale-api-secret-key")
	d.InstanceProfile = flags.String("exoscale-instance-profile")
	d.DiskSize = int64(flags.Int("exoscale-disk-size"))
	d.Image = flags.String("exoscale-image")
	securityGroups := flags.StringSlice("exoscale-security-group")
	if len(securityGroups) == 0 {
		securityGroups = []string{"docker-machine"}
	}
	d.SecurityGroup = strings.Join(securityGroups, ",")
	affinityGroups := flags.StringSlice("exoscale-affinity-group")
	if len(affinityGroups) > 0 {
		d.AffinityGroup = strings.Join(affinityGroups, ",")
	}
	d.AvailabilityZone = flags.String("exoscale-availability-zone")
	d.SSHUser = flags.String("exoscale-ssh-user")
	d.UserDataFile = flags.String("exoscale-userdata")
	d.SetSwarmConfigFromFlags(flags)

	if d.URL == "" {
		d.URL = defaultAPIEndpoint
	}
	if d.APIKey == "" || d.APISecretKey == "" {
		return errors.New("missing an API key (--exoscale-api-key) or API secret key (--exoscale-api-secret-key)")
	}

	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is
// ready for creation
func (d *Driver) PreCreateCheck() error {
	if d.UserDataFile != "" {
		if _, err := os.Stat(d.UserDataFile); os.IsNotExist(err) {
			return fmt.Errorf("user-data file %s could not be found", d.UserDataFile)
		}
	}

	return nil
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g tcp://10.1.2.3:2376
func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) client() *egoscale.Client {
	return egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)
}

func (d *Driver) virtualMachine() (*egoscale.VirtualMachine, error) {
	cs := d.client()
	resp, err := cs.Request(&egoscale.ListVirtualMachines{
		ID: d.ID,
	})
	if err != nil {
		return nil, err
	}
	vms := resp.(*egoscale.ListVirtualMachinesResponse)
	if vms.Count == 0 {
		return nil, fmt.Errorf("No VM found. %s", d.ID)
	}

	virtualMachine := vms.VirtualMachine[0]
	return &virtualMachine, nil
}

// GetState returns a github.com/machine/libmachine/state.State representing the state of the host (running, stopped, etc.)
func (d *Driver) GetState() (state.State, error) {
	vm, err := d.virtualMachine()
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

func (d *Driver) createDefaultSecurityGroup(group string) (*egoscale.SecurityGroup, error) {
	cs := d.client()
	resp, err := cs.Request(&egoscale.CreateSecurityGroup{
		Name:        group,
		Description: "created by docker-machine",
	})
	if err != nil {
		return nil, err
	}
	sg := resp.(*egoscale.CreateSecurityGroupResponse).SecurityGroup

	requests := []egoscale.AuthorizeSecurityGroupIngress{
		{
			SecurityGroupID: sg.ID,
			Description:     "SSH",
			CidrList:        []string{"0.0.0.0/0"},
			Protocol:        "TCP",
			StartPort:       22,
			EndPort:         22,
		},
		{
			SecurityGroupID: sg.ID,
			Description:     "Ping",
			CidrList:        []string{"0.0.0.0/0"},
			Protocol:        "ICMP",
			IcmpType:        8,
			IcmpCode:        0,
		},
		{
			SecurityGroupID: sg.ID,
			Description:     "Docker",
			CidrList:        []string{"0.0.0.0/0"},
			Protocol:        "TCP",
			StartPort:       2376,
			EndPort:         2377,
		},
		{
			SecurityGroupID: sg.ID,
			Description:     "Legacy Standalone Swarm",
			CidrList:        []string{"0.0.0.0/0"},
			Protocol:        "TCP",
			StartPort:       3376,
			EndPort:         3377,
		},
		{
			SecurityGroupID: sg.ID,
			Description:     "Communication among nodes",
			Protocol:        "TCP",
			StartPort:       7946,
			EndPort:         7946,
			UserSecurityGroupList: []egoscale.UserSecurityGroup{{
				Group:   sg.Name,
				Account: sg.Account,
			}},
		},
		{
			SecurityGroupID: sg.ID,
			Description:     "Communication among nodes",
			Protocol:        "UDP",
			StartPort:       7946,
			EndPort:         7946,
			UserSecurityGroupList: []egoscale.UserSecurityGroup{{
				Group:   sg.Name,
				Account: sg.Account,
			}},
		},
		{
			SecurityGroupID: sg.ID,
			Description:     "Overlay network traffic",
			Protocol:        "UDP",
			StartPort:       4789,
			EndPort:         4789,
			UserSecurityGroupList: []egoscale.UserSecurityGroup{{
				Group:   sg.Name,
				Account: sg.Account,
			}},
		},
	}

	for _, req := range requests {
		_, err := cs.AsyncRequest(&req, d.async)
		if err != nil {
			return nil, err
		}
	}

	return &sg, nil
}

func (d *Driver) createDefaultAffinityGroup(group string) (*egoscale.AffinityGroup, error) {
	cs := d.client()
	resp, err := cs.AsyncRequest(&egoscale.CreateAffinityGroup{
		Name:        group,
		Type:        defaultAffinityGroupType,
		Description: "created by docker-machine",
	}, d.async)

	if err != nil {
		return nil, err
	}

	affinityGroup := resp.(*egoscale.CreateAffinityGroupResponse).AffinityGroup
	return &affinityGroup, nil
}

// Create creates the VM instance acting as the docker host
func (d *Driver) Create() error {
	cloudInit, err := d.getCloudInit()
	if err != nil {
		return err
	}
	userData := base64.StdEncoding.EncodeToString(cloudInit)

	log.Infof("Querying exoscale for the requested parameters...")
	client := egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)
	topology, err := client.GetTopology()
	if err != nil {
		return err
	}

	// Availability zone UUID
	zone, ok := topology.Zones[strings.ToLower(d.AvailabilityZone)]
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

	// Security groups
	securityGroups := strings.Split(d.SecurityGroup, ",")
	sgs := make([]string, len(securityGroups))
	for idx, group := range securityGroups {
		sg, ok := topology.SecurityGroups[group]
		if !ok {
			log.Infof("Security group %v does not exist, create it", group)
			securityGroup, err := d.createDefaultSecurityGroup(group)
			if err != nil {
				return err
			}
			sg = securityGroup.ID
		}
		log.Debugf("Security group %v = %s", group, sg)
		sgs[idx] = sg
	}

	// Affinity Groups
	affinityGroups := strings.Split(d.AffinityGroup, ",")
	ags := make([]string, len(affinityGroups))
	for idx, group := range affinityGroups {
		ag, ok := topology.AffinityGroups[group]
		if !ok {
			log.Infof("Affinity Group %v does not exist, create it", group)
			affinityGroup, err := d.createDefaultAffinityGroup(group)
			if err != nil {
				return err
			}
			ag = affinityGroup.ID
		}
		log.Debugf("Affinity group %v = %s", group, ag)
		ags[idx] = ag
	}

	log.Infof("Generate an SSH keypair...")
	keypairName := fmt.Sprintf("docker-machine-%s", d.MachineName)
	kpresp, err := client.CreateKeypair(keypairName)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(d.GetSSHKeyPath(), []byte(kpresp.PrivateKey), 0600)
	if err != nil {
		return err
	}
	d.KeyPair = keypairName

	log.Infof("Spawn exoscale host...")
	log.Debugf("Using the following cloud-init file:")
	log.Debugf("%s", string(cloudInit))

	req := &egoscale.DeployVirtualMachine{
		TemplateID:        tpl,
		ServiceOfferingID: profile,
		UserData:          userData,
		ZoneID:            zone,
		KeyPair:           d.KeyPair,
		Name:              d.MachineName,
		DisplayName:       d.MachineName,
		RootDiskSize:      d.DiskSize,
		SecurityGroupIDs:  sgs,
		AffinityGroupIDs:  ags,
	}
	log.Infof("Deploy %#v", req)
	resp, err := client.AsyncRequest(req, d.async)
	if err != nil {
		return err
	}

	vm := resp.(*egoscale.DeployVirtualMachineResponse).VirtualMachine

	IPAddress := vm.Nic[0].IPAddress
	if IPAddress != nil {
		d.IPAddress = IPAddress.String()
	}
	d.ID = vm.ID

	return nil
}

// Start starts the existing VM instance.
func (d *Driver) Start() error {
	cs := d.client()
	_, err := cs.AsyncRequest(&egoscale.StartVirtualMachine{
		ID: d.ID,
	}, d.async)

	return err
}

// Stop stops the existing VM instance.
func (d *Driver) Stop() error {
	cs := d.client()
	_, err := cs.AsyncRequest(&egoscale.StopVirtualMachine{
		ID: d.ID,
	}, d.async)

	return err
}

// Restart reboots the existing VM instance.
func (d *Driver) Restart() error {
	cs := d.client()
	_, err := cs.AsyncRequest(&egoscale.RebootVirtualMachine{
		ID: d.ID,
	}, d.async)

	return err
}

// Kill stops a host forcefully (same as Stop)
func (d *Driver) Kill() error {
	return d.Stop()
}

// Remove destroys the VM instance and the associated SSH key.
func (d *Driver) Remove() error {
	cs := d.client()

	// Destroy the SSH key
	if err := cs.BooleanRequest(&egoscale.DeleteSSHKeyPair{Name: d.KeyPair}); err != nil {
		return err
	}

	// Destroy the virtual machine
	_, err := cs.AsyncRequest(&egoscale.DestroyVirtualMachine{ID: d.ID}, d.async)

	log.Infof("The Anti-Affinity group and Security group were not removed")

	return err
}

// Build a cloud-init user data string that will install and run
// docker.
func (d *Driver) getCloudInit() ([]byte, error) {
	if d.UserDataFile != "" {
		return ioutil.ReadFile(d.UserDataFile)
	}

	return []byte(defaultCloudInit), nil
}
