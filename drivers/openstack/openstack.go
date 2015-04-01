package openstack

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	dockerConfigDir = "/etc/docker"
)

type Driver struct {
	AuthURL          string
	Insecure         bool
	Username         string
	Password         string
	TenantName       string
	TenantID         string
	Region           string
	EndpointType     string
	MachineName      string
	MachineID        string
	FlavorName       string
	FlavorID         string
	ImageName        string
	ImageID          string
	KeyPairName      string
	NetworkName      string
	NetworkID        string
	SecurityGroups   []string
	FloatingIPPool   string
	FloatingIPPoolID string
	SSHUser          string
	SSHPort          int
	IP               string
	CaCertPath       string
	PrivateKeyPath   string
	storePath        string
	SwarmMaster      bool
	SwarmHost        string
	SwarmDiscovery   string
	client           Client
}

func init() {
	drivers.Register("openstack", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "OS_AUTH_URL",
			Name:   "openstack-auth-url",
			Usage:  "OpenStack authentication URL",
			Value:  "",
		},
		cli.BoolFlag{
			Name:  "openstack-insecure",
			Usage: "Disable TLS credential checking.",
		},
		cli.StringFlag{
			EnvVar: "OS_USERNAME",
			Name:   "openstack-username",
			Usage:  "OpenStack username",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_PASSWORD",
			Name:   "openstack-password",
			Usage:  "OpenStack password",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_TENANT_NAME",
			Name:   "openstack-tenant-name",
			Usage:  "OpenStack tenant name",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_TENANT_ID",
			Name:   "openstack-tenant-id",
			Usage:  "OpenStack tenant id",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_REGION_NAME",
			Name:   "openstack-region",
			Usage:  "OpenStack region name",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_ENDPOINT_TYPE",
			Name:   "openstack-endpoint-type",
			Usage:  "OpenStack endpoint type (adminURL, internalURL or publicURL)",
			Value:  "",
		},
		cli.StringFlag{
			Name:  "openstack-flavor-id",
			Usage: "OpenStack flavor id to use for the instance",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-flavor-name",
			Usage: "OpenStack flavor name to use for the instance",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-image-id",
			Usage: "OpenStack image id to use for the instance",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-image-name",
			Usage: "OpenStack image name to use for the instance",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-net-id",
			Usage: "OpenStack network id the machine will be connected on",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-net-name",
			Usage: "OpenStack network name the machine will be connected on",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-sec-groups",
			Usage: "OpenStack comma separated security groups for the machine",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-floatingip-pool",
			Usage: "OpenStack floating IP pool to get an IP from to assign to the instance",
			Value: "",
		},
		cli.StringFlag{
			Name:  "openstack-ssh-user",
			Usage: "OpenStack SSH user",
			Value: "root",
		},
		cli.IntFlag{
			Name:  "openstack-ssh-port",
			Usage: "OpenStack SSH port",
			Value: 22,
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	log.WithFields(log.Fields{
		"machineName": machineName,
		"storePath":   storePath,
		"caCert":      caCert,
		"privateKey":  privateKey,
	}).Debug("Instantiating OpenStack driver...")

	return NewDerivedDriver(machineName, storePath, &GenericClient{}, caCert, privateKey)
}

func NewDerivedDriver(machineName string, storePath string, client Client, caCert string, privateKey string) (*Driver, error) {
	return &Driver{
		MachineName:    machineName,
		storePath:      storePath,
		client:         client,
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
	return "openstack"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AuthURL = flags.String("openstack-auth-url")
	d.Insecure = flags.Bool("openstack-insecure")
	d.Username = flags.String("openstack-username")
	d.Password = flags.String("openstack-password")
	d.TenantName = flags.String("openstack-tenant-name")
	d.TenantID = flags.String("openstack-tenant-id")
	d.Region = flags.String("openstack-region")
	d.EndpointType = flags.String("openstack-endpoint-type")
	d.FlavorID = flags.String("openstack-flavor-id")
	d.FlavorName = flags.String("openstack-flavor-name")
	d.ImageID = flags.String("openstack-image-id")
	d.ImageName = flags.String("openstack-image-name")
	d.NetworkID = flags.String("openstack-net-id")
	d.NetworkName = flags.String("openstack-net-name")
	if flags.String("openstack-sec-groups") != "" {
		d.SecurityGroups = strings.Split(flags.String("openstack-sec-groups"), ",")
	}
	d.FloatingIPPool = flags.String("openstack-floatingip-pool")
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
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IP != "" {
		return d.IP, nil
	}

	log.WithField("MachineID", d.MachineID).Debug("Looking for the IP address...")

	if err := d.initCompute(); err != nil {
		return "", err
	}

	addressType := Fixed
	if d.FloatingIPPool != "" {
		addressType = Floating
	}

	// Looking for the IP address in a retry loop to deal with OpenStack latency
	for retryCount := 0; retryCount < 200; retryCount++ {
		addresses, err := d.client.GetInstanceIPAddresses(d)
		if err != nil {
			return "", err
		}
		for _, a := range addresses {
			if a.AddressType == addressType {
				return a.Address, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("No IP found for the machine")
}

func (d *Driver) GetState() (state.State, error) {
	log.WithField("MachineID", d.MachineID).Debug("Get status for OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return state.None, err
	}

	s, err := d.client.GetInstanceState(d)
	if err != nil {
		return state.None, err
	}

	log.WithFields(log.Fields{
		"MachineID": d.MachineID,
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

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	d.KeyPairName = fmt.Sprintf("%s-%s", d.MachineName, utils.GenerateRandomID())

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
	if d.FloatingIPPool != "" {
		if err := d.assignFloatingIP(); err != nil {
			return err
		}
	}
	if err := d.lookForIPAddress(); err != nil {
		return err
	}
	if err := d.waitForSSHServer(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Start() error {
	log.WithField("MachineID", d.MachineID).Info("Starting OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.StartInstance(d); err != nil {
		return err
	}
	return d.waitForInstanceToStart()
}

func (d *Driver) Stop() error {
	log.WithField("MachineID", d.MachineID).Info("Stopping OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.StopInstance(d); err != nil {
		return err
	}

	log.WithField("MachineID", d.MachineID).Info("Waiting for the OpenStack instance to stop...")
	if err := d.client.WaitForInstanceStatus(d, "SHUTOFF", 200); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Remove() error {
	log.WithField("MachineID", d.MachineID).Debug("deleting instance...")
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
	log.WithField("MachineID", d.MachineID).Info("Restarting OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.RestartInstance(d); err != nil {
		return err
	}
	return d.waitForInstanceToStart()
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
	if d.AuthURL == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Authentication URL", "OS_AUTH_URL", "--openstack-auth-url")
	}
	if d.Username == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Username", "OS_USERNAME", "--openstack-username")
	}
	if d.Password == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Password", "OS_PASSWORD", "--openstack-password")
	}
	if d.TenantName == "" && d.TenantID == "" {
		return fmt.Errorf(errorMandatoryTenantNameOrID)
	}

	if d.FlavorName == "" && d.FlavorID == "" {
		return fmt.Errorf(errorMandatoryOption, "Flavor name or Flavor id", "--openstack-flavor-name or --openstack-flavor-id")
	}
	if d.FlavorName != "" && d.FlavorID != "" {
		return fmt.Errorf(errorExclusiveOptions, "Flavor name", "Flavor id")
	}

	if d.ImageName == "" && d.ImageID == "" {
		return fmt.Errorf(errorMandatoryOption, "Image name or Image id", "--openstack-image-name or --openstack-image-id")
	}
	if d.ImageName != "" && d.ImageID != "" {
		return fmt.Errorf(errorExclusiveOptions, "Image name", "Image id")
	}

	if d.NetworkName != "" && d.NetworkID != "" {
		return fmt.Errorf(errorExclusiveOptions, "Network name", "Network id")
	}
	if d.EndpointType != "" && (d.EndpointType != "publicURL" && d.EndpointType != "adminURL" && d.EndpointType != "internalURL") {
		return fmt.Errorf(errorWrongEndpointType)
	}
	return nil
}

func (d *Driver) resolveIds() error {
	if d.NetworkName != "" {
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

		d.NetworkID = networkID
		log.WithFields(log.Fields{
			"Name": d.NetworkName,
			"ID":   d.NetworkID,
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

		d.FlavorID = flavorID
		log.WithFields(log.Fields{
			"Name": d.FlavorName,
			"ID":   d.FlavorID,
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

		d.ImageID = imageID
		log.WithFields(log.Fields{
			"Name": d.ImageName,
			"ID":   d.ImageID,
		}).Debug("Found image id using its name")
	}

	if d.FloatingIPPool != "" {
		if err := d.initNetwork(); err != nil {
			return err
		}
		f, err := d.client.GetFloatingIPPoolID(d)

		if err != nil {
			return err
		}

		if f == "" {
			return fmt.Errorf(errorUnknownNetworkName, d.FloatingIPPool)
		}

		d.FloatingIPPoolID = f
		log.WithFields(log.Fields{
			"Name": d.FloatingIPPool,
			"ID":   d.FloatingIPPoolID,
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
		"FlavorID": d.FlavorID,
		"ImageID":  d.ImageID,
	}).Debug("Creating OpenStack instance...")

	if err := d.initCompute(); err != nil {
		return err
	}
	instanceID, err := d.client.CreateInstance(d)
	if err != nil {
		return err
	}
	d.MachineID = instanceID
	return nil
}

func (d *Driver) assignFloatingIP() error {

	if err := d.initNetwork(); err != nil {
		return err
	}

	portID, err := d.client.GetInstancePortID(d)
	if err != nil {
		return err
	}

	ips, err := d.client.GetFloatingIPs(d)
	if err != nil {
		return err
	}

	var floatingIP *FloatingIP

	log.WithFields(log.Fields{
		"MachineID": d.MachineID,
		"Pool":      d.FloatingIPPool,
	}).Debugf("Looking for an available floating IP")

	for _, ip := range ips {
		if ip.PortID == "" {
			log.WithFields(log.Fields{
				"MachineID": d.MachineID,
				"IP":        ip.IP,
			}).Debugf("Available floating IP found")
			floatingIP = &ip
			break
		}
	}

	if floatingIP == nil {
		floatingIP = &FloatingIP{}
		log.WithField("MachineID", d.MachineID).Debugf("No available floating IP found. Allocating a new one...")
	} else {
		log.WithField("MachineID", d.MachineID).Debugf("Assigning floating IP to the instance")
	}

	if err := d.client.AssignFloatingIP(d, floatingIP, portID); err != nil {
		return err
	}
	d.IP = floatingIP.IP
	return nil
}

func (d *Driver) waitForInstanceActive() error {
	log.WithField("MachineID", d.MachineID).Debug("Waiting for the OpenStack instance to be ACTIVE...")
	if err := d.client.WaitForInstanceStatus(d, "ACTIVE", 200); err != nil {
		return err
	}
	return nil
}

func (d *Driver) lookForIPAddress() error {
	ip, err := d.GetIP()
	if err != nil {
		return err
	}
	d.IP = ip
	log.WithFields(log.Fields{
		"IP":        ip,
		"MachineID": d.MachineID,
	}).Debug("IP address found")
	return nil
}

func (d *Driver) waitForSSHServer() error {
	ip, err := d.GetIP()
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"MachineID": d.MachineID,
		"IP":        ip,
	}).Debug("Waiting for the SSH server to be started...")
	return ssh.WaitForTCP(fmt.Sprintf("%s:%d", ip, d.SSHPort))
}

func (d *Driver) waitForInstanceToStart() error {
	if err := d.waitForInstanceActive(); err != nil {
		return err
	}
	return d.waitForSSHServer()
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}
