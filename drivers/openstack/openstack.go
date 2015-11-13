package openstack

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	AuthUrl          string
	ActiveTimeout    int
	Insecure         bool
	DomainID         string
	DomainName       string
	Username         string
	Password         string
	TenantName       string
	TenantId         string
	Region           string
	AvailabilityZone string
	EndpointType     string
	MachineId        string
	FlavorName       string
	FlavorId         string
	ImageName        string
	ImageId          string
	KeyPairName      string
	NetworkName      string
	NetworkId        string
	SecurityGroups   []string
	FloatingIpPool   string
	ComputeNetwork   bool
	FloatingIpPoolId string
	IpVersion        int
	client           Client
}

const (
	defaultSSHUser       = "root"
	defaultSSHPort       = 22
	defaultActiveTimeout = 200
)

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "OS_AUTH_URL",
			Name:   "openstack-auth-url",
			Usage:  "OpenStack authentication URL",
			Value:  "",
		},
		mcnflag.BoolFlag{
			EnvVar: "OS_INSECURE",
			Name:   "openstack-insecure",
			Usage:  "Disable TLS credential checking.",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_DOMAIN_ID",
			Name:   "openstack-domain-id",
			Usage:  "OpenStack domain ID (identity v3 only)",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_DOMAIN_NAME",
			Name:   "openstack-domain-name",
			Usage:  "OpenStack domain name (identity v3 only)",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_USERNAME",
			Name:   "openstack-username",
			Usage:  "OpenStack username",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_PASSWORD",
			Name:   "openstack-password",
			Usage:  "OpenStack password",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_TENANT_NAME",
			Name:   "openstack-tenant-name",
			Usage:  "OpenStack tenant name",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_TENANT_ID",
			Name:   "openstack-tenant-id",
			Usage:  "OpenStack tenant id",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_REGION_NAME",
			Name:   "openstack-region",
			Usage:  "OpenStack region name",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_AVAILABILITY_ZONE",
			Name:   "openstack-availability-zone",
			Usage:  "OpenStack availability zone",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_ENDPOINT_TYPE",
			Name:   "openstack-endpoint-type",
			Usage:  "OpenStack endpoint type (adminURL, internalURL or publicURL)",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_FLAVOR_ID",
			Name:   "openstack-flavor-id",
			Usage:  "OpenStack flavor id to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_FLAVOR_NAME",
			Name:   "openstack-flavor-name",
			Usage:  "OpenStack flavor name to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_IMAGE_ID",
			Name:   "openstack-image-id",
			Usage:  "OpenStack image id to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_IMAGE_NAME",
			Name:   "openstack-image-name",
			Usage:  "OpenStack image name to use for the instance",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_NETWORK_ID",
			Name:   "openstack-net-id",
			Usage:  "OpenStack network id the machine will be connected on",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_NETWORK_NAME",
			Name:   "openstack-net-name",
			Usage:  "OpenStack network name the machine will be connected on",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_SECURITY_GROUPS",
			Name:   "openstack-sec-groups",
			Usage:  "OpenStack comma separated security groups for the machine",
			Value:  "",
		},
		mcnflag.BoolFlag{
			EnvVar: "OS_NOVA_NETWORK",
			Name:   "openstack-nova-network",
			Usage:  "Use the nova networking services instead of neutron.",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_FLOATINGIP_POOL",
			Name:   "openstack-floatingip-pool",
			Usage:  "OpenStack floating IP pool to get an IP from to assign to the instance",
			Value:  "",
		},
		mcnflag.IntFlag{
			EnvVar: "OS_IP_VERSION",
			Name:   "openstack-ip-version",
			Usage:  "OpenStack version of IP address assigned for the machine",
			Value:  4,
		},
		mcnflag.StringFlag{
			EnvVar: "OS_SSH_USER",
			Name:   "openstack-ssh-user",
			Usage:  "OpenStack SSH user",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			EnvVar: "OS_SSH_PORT",
			Name:   "openstack-ssh-port",
			Usage:  "OpenStack SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.IntFlag{
			EnvVar: "OS_ACTIVE_TIMEOUT",
			Name:   "openstack-active-timeout",
			Usage:  "OpenStack active timeout",
			Value:  defaultActiveTimeout,
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return NewDerivedDriver(hostName, storePath)
}

func NewDerivedDriver(hostName, storePath string) *Driver {
	return &Driver{
		client:        &GenericClient{},
		ActiveTimeout: defaultActiveTimeout,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			SSHPort:     defaultSSHPort,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) SetClient(client Client) {
	d.client = client
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "openstack"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AuthUrl = flags.String("openstack-auth-url")
	d.ActiveTimeout = flags.Int("openstack-active-timeout")
	d.Insecure = flags.Bool("openstack-insecure")
	d.DomainID = flags.String("openstack-domain-id")
	d.DomainName = flags.String("openstack-domain-name")
	d.Username = flags.String("openstack-username")
	d.Password = flags.String("openstack-password")
	d.TenantName = flags.String("openstack-tenant-name")
	d.TenantId = flags.String("openstack-tenant-id")
	d.Region = flags.String("openstack-region")
	d.AvailabilityZone = flags.String("openstack-availability-zone")
	d.EndpointType = flags.String("openstack-endpoint-type")
	d.FlavorId = flags.String("openstack-flavor-id")
	d.FlavorName = flags.String("openstack-flavor-name")
	d.ImageId = flags.String("openstack-image-id")
	d.ImageName = flags.String("openstack-image-name")
	d.NetworkId = flags.String("openstack-net-id")
	d.NetworkName = flags.String("openstack-net-name")
	if flags.String("openstack-sec-groups") != "" {
		d.SecurityGroups = strings.Split(flags.String("openstack-sec-groups"), ",")
	}
	d.FloatingIpPool = flags.String("openstack-floatingip-pool")
	d.IpVersion = flags.Int("openstack-ip-version")
	d.ComputeNetwork = flags.Bool("openstack-nova-network")
	d.SSHUser = flags.String("openstack-ssh-user")
	d.SSHPort = flags.Int("openstack-ssh-port")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	return d.checkConfig()
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress != "" {
		return d.IPAddress, nil
	}

	log.WithField("MachineId", d.MachineId).Debug("Looking for the IP address...")

	if err := d.initCompute(); err != nil {
		return "", err
	}

	addressType := Fixed
	if d.FloatingIpPool != "" {
		addressType = Floating
	}

	// Looking for the IP address in a retry loop to deal with OpenStack latency
	for retryCount := 0; retryCount < 200; retryCount++ {
		addresses, err := d.client.GetInstanceIPAddresses(d)
		if err != nil {
			return "", err
		}
		for _, a := range addresses {
			if a.AddressType == addressType && a.Version == d.IpVersion {
				return a.Address, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("No IP found for the machine")
}

func (d *Driver) GetState() (state.State, error) {
	log.WithField("MachineId", d.MachineId).Debug("Get status for OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return state.None, err
	}

	s, err := d.client.GetInstanceState(d)
	if err != nil {
		return state.None, err
	}

	log.WithFields(log.Fields{
		"MachineId": d.MachineId,
		"State":     s,
	}).Debug("State for OpenStack instance")

	switch s {
	case "ACTIVE":
		return state.Running, nil
	case "PAUSED":
		return state.Paused, nil
	case "SUSPENDED":
		return state.Saved, nil
	case "SHUTOFF":
		return state.Stopped, nil
	case "BUILDING":
		return state.Starting, nil
	case "ERROR":
		return state.Error, nil
	}
	return state.None, nil
}

func (d *Driver) Create() error {
	d.KeyPairName = fmt.Sprintf("%s-%s", d.MachineName, mcnutils.GenerateRandomID())

	if err := d.resolveIds(); err != nil {
		return err
	}
	if err := d.createSSHKey(); err != nil {
		return err
	}
	if err := d.createMachine(); err != nil {
		return err
	}
	if err := d.waitForInstanceActive(); err != nil {
		return err
	}
	if d.FloatingIpPool != "" {
		if err := d.assignFloatingIP(); err != nil {
			return err
		}
	}
	if err := d.lookForIPAddress(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Start() error {
	log.WithField("MachineId", d.MachineId).Info("Starting OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.StartInstance(d); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Stop() error {
	log.WithField("MachineId", d.MachineId).Info("Stopping OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.StopInstance(d); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Remove() error {
	log.WithField("MachineId", d.MachineId).Debug("deleting instance...")
	log.Info("Deleting OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.DeleteInstance(d); err != nil {
		return err
	}
	log.WithField("Name", d.KeyPairName).Debug("deleting key pair...")
	if err := d.client.DeleteKeyPair(d, d.KeyPairName); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	log.WithField("MachineId", d.MachineId).Info("Restarting OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.RestartInstance(d); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

const (
	errorMandatoryEnvOrOption    string = "%s must be specified either using the environment variable %s or the CLI option %s"
	errorMandatoryOption         string = "%s must be specified using the CLI option %s"
	errorExclusiveOptions        string = "Either %s or %s must be specified, not both"
	errorMandatoryTenantNameOrID string = "Tenant id or name must be provided either using one of the environment variables OS_TENANT_ID and OS_TENANT_NAME or one of the CLI options --openstack-tenant-id and --openstack-tenant-name"
	errorWrongEndpointType       string = "Endpoint type must be 'publicURL', 'adminURL' or 'internalURL'"
	errorUnknownFlavorName       string = "Unable to find flavor named %s"
	errorUnknownImageName        string = "Unable to find image named %s"
	errorUnknownNetworkName      string = "Unable to find network named %s"
)

func (d *Driver) checkConfig() error {
	if d.AuthUrl == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Authentication URL", "OS_AUTH_URL", "--openstack-auth-url")
	}
	if d.Username == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Username", "OS_USERNAME", "--openstack-username")
	}
	if d.Password == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Password", "OS_PASSWORD", "--openstack-password")
	}
	if d.TenantName == "" && d.TenantId == "" {
		return fmt.Errorf(errorMandatoryTenantNameOrID)
	}

	if d.FlavorName == "" && d.FlavorId == "" {
		return fmt.Errorf(errorMandatoryOption, "Flavor name or Flavor id", "--openstack-flavor-name or --openstack-flavor-id")
	}
	if d.FlavorName != "" && d.FlavorId != "" {
		return fmt.Errorf(errorExclusiveOptions, "Flavor name", "Flavor id")
	}

	if d.ImageName == "" && d.ImageId == "" {
		return fmt.Errorf(errorMandatoryOption, "Image name or Image id", "--openstack-image-name or --openstack-image-id")
	}
	if d.ImageName != "" && d.ImageId != "" {
		return fmt.Errorf(errorExclusiveOptions, "Image name", "Image id")
	}

	if d.NetworkName != "" && d.NetworkId != "" {
		return fmt.Errorf(errorExclusiveOptions, "Network name", "Network id")
	}
	if d.EndpointType != "" && (d.EndpointType != "publicURL" && d.EndpointType != "adminURL" && d.EndpointType != "internalURL") {
		return fmt.Errorf(errorWrongEndpointType)
	}
	return nil
}

func (d *Driver) resolveIds() error {
	if d.NetworkName != "" && !d.ComputeNetwork {
		if err := d.initNetwork(); err != nil {
			return err
		}

		networkID, err := d.client.GetNetworkID(d)

		if err != nil {
			return err
		}

		if networkID == "" {
			return fmt.Errorf(errorUnknownNetworkName, d.NetworkName)
		}

		d.NetworkId = networkID
		log.WithFields(log.Fields{
			"Name": d.NetworkName,
			"ID":   d.NetworkId,
		}).Debug("Found network id using its name")
	}

	if d.FlavorName != "" {
		if err := d.initCompute(); err != nil {
			return err
		}
		flavorID, err := d.client.GetFlavorID(d)

		if err != nil {
			return err
		}

		if flavorID == "" {
			return fmt.Errorf(errorUnknownFlavorName, d.FlavorName)
		}

		d.FlavorId = flavorID
		log.WithFields(log.Fields{
			"Name": d.FlavorName,
			"ID":   d.FlavorId,
		}).Debug("Found flavor id using its name")
	}

	if d.ImageName != "" {
		if err := d.initCompute(); err != nil {
			return err
		}
		imageID, err := d.client.GetImageID(d)

		if err != nil {
			return err
		}

		if imageID == "" {
			return fmt.Errorf(errorUnknownImageName, d.ImageName)
		}

		d.ImageId = imageID
		log.WithFields(log.Fields{
			"Name": d.ImageName,
			"ID":   d.ImageId,
		}).Debug("Found image id using its name")
	}

	if d.FloatingIpPool != "" && !d.ComputeNetwork {
		if err := d.initNetwork(); err != nil {
			return err
		}
		f, err := d.client.GetFloatingIPPoolID(d)

		if err != nil {
			return err
		}

		if f == "" {
			return fmt.Errorf(errorUnknownNetworkName, d.FloatingIpPool)
		}

		d.FloatingIpPoolId = f
		log.WithFields(log.Fields{
			"Name": d.FloatingIpPool,
			"ID":   d.FloatingIpPoolId,
		}).Debug("Found floating IP pool id using its name")
	}

	return nil
}

func (d *Driver) initCompute() error {
	if err := d.client.Authenticate(d); err != nil {
		return err
	}
	if err := d.client.InitComputeClient(d); err != nil {
		return err
	}
	return nil
}

func (d *Driver) initNetwork() error {
	if err := d.client.Authenticate(d); err != nil {
		return err
	}
	if err := d.client.InitNetworkClient(d); err != nil {
		return err
	}
	return nil
}

func (d *Driver) createSSHKey() error {
	log.WithField("Name", d.KeyPairName).Debug("Creating Key Pair...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}
	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.CreateKeyPair(d, d.KeyPairName, string(publicKey)); err != nil {
		return err
	}
	return nil
}

func (d *Driver) createMachine() error {
	log.WithFields(log.Fields{
		"FlavorId": d.FlavorId,
		"ImageId":  d.ImageId,
	}).Debug("Creating OpenStack instance...")

	if err := d.initCompute(); err != nil {
		return err
	}
	instanceID, err := d.client.CreateInstance(d)
	if err != nil {
		return err
	}
	d.MachineId = instanceID
	return nil
}

func (d *Driver) assignFloatingIP() error {
	var err error

	if d.ComputeNetwork {
		err = d.initCompute()
	} else {
		err = d.initNetwork()
	}

	if err != nil {
		return err
	}

	ips, err := d.client.GetFloatingIPs(d)
	if err != nil {
		return err
	}

	var floatingIP *FloatingIP

	log.WithFields(log.Fields{
		"MachineId": d.MachineId,
		"Pool":      d.FloatingIpPool,
	}).Debugf("Looking for an available floating IP")

	for _, ip := range ips {
		if ip.PortId == "" {
			log.WithFields(log.Fields{
				"MachineId": d.MachineId,
				"IP":        ip.Ip,
			}).Debugf("Available floating IP found")
			floatingIP = &ip
			break
		}
	}

	if floatingIP == nil {
		floatingIP = &FloatingIP{}
		log.WithField("MachineId", d.MachineId).Debugf("No available floating IP found. Allocating a new one...")
	} else {
		log.WithField("MachineId", d.MachineId).Debugf("Assigning floating IP to the instance")
	}

	if err := d.client.AssignFloatingIP(d, floatingIP); err != nil {
		return err
	}
	d.IPAddress = floatingIP.Ip
	return nil
}

func (d *Driver) waitForInstanceActive() error {
	log.WithField("MachineId", d.MachineId).Debug("Waiting for the OpenStack instance to be ACTIVE...")
	if err := d.client.WaitForInstanceStatus(d, "ACTIVE"); err != nil {
		return err
	}
	return nil
}

func (d *Driver) lookForIPAddress() error {
	ip, err := d.GetIP()
	if err != nil {
		return err
	}
	d.IPAddress = ip
	log.WithFields(log.Fields{
		"IP":        ip,
		"MachineId": d.MachineId,
	}).Debug("IP address found")
	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}
