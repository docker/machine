package openstack

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	AuthUrl        string
	Username       string
	Password       string
	TenantName     string
	TenantId       string
	Region         string
	EndpointType   string
	MachineName    string
	MachineId      string
	FlavorName     string
	FlavorId       string
	ImageName      string
	ImageId        string
	KeyPairName    string
	NetworkName    string
	NetworkId      string
	SecurityGroups []string
	FloatingIpPool string
	SSHUser        string
	SSHPort        int
	storePath      string
	client         Client
}

type CreateFlags struct {
	AuthUrl        *string
	Username       *string
	Password       *string
	TenantName     *string
	TenantId       *string
	Region         *string
	EndpointType   *string
	FlavorName     *string
	FlavorId       *string
	ImageName      *string
	ImageId        *string
	NetworkName    *string
	NetworkId      *string
	SecurityGroups *string
	FloatingIpPool *string
	SSHUser        *string
	SSHPort        *int
}

func init() {
	drivers.Register("openstack", &drivers.RegisteredDriver{
		New:                 NewDriver,
		RegisterCreateFlags: RegisterCreateFlags,
	})
}

func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
	createFlags := new(CreateFlags)
	createFlags.AuthUrl = cmd.String(
		[]string{"-openstack-auth-url"},
		os.Getenv("OS_AUTH_URL"),
		"OpenStack authentication URL",
	)
	createFlags.Username = cmd.String(
		[]string{"-openstack-username"},
		os.Getenv("OS_USERNAME"),
		"OpenStack username",
	)
	createFlags.Password = cmd.String(
		[]string{"-openstack-password"},
		os.Getenv("OS_PASSWORD"),
		"OpenStack password",
	)
	createFlags.TenantName = cmd.String(
		[]string{"-openstack-tenant-name"},
		os.Getenv("OS_TENANT_NAME"),
		"OpenStack tenant name",
	)
	createFlags.TenantId = cmd.String(
		[]string{"-openstack-tenant-id"},
		os.Getenv("OS_TENANT_ID"),
		"OpenStack tenant id",
	)
	createFlags.Region = cmd.String(
		[]string{"-openstack-region"},
		os.Getenv("OS_REGION_NAME"),
		"OpenStack region name",
	)
	createFlags.EndpointType = cmd.String(
		[]string{"-openstack-endpoint-type"},
		os.Getenv("OS_ENDPOINT_TYPE"),
		"OpenStack endpoint type (adminURL, internalURL or the default publicURL)",
	)
	createFlags.FlavorId = cmd.String(
		[]string{"-openstack-flavor-id"},
		"",
		"OpenStack flavor id to use for the instance",
	)
	createFlags.FlavorName = cmd.String(
		[]string{"-openstack-flavor-name"},
		"",
		"OpenStack flavor name to use for the instance",
	)
	createFlags.ImageId = cmd.String(
		[]string{"-openstack-image-id"},
		"",
		"OpenStack image id to use for the instance",
	)
	createFlags.ImageName = cmd.String(
		[]string{"-openstack-image-name"},
		"",
		"OpenStack image name to use for the instance",
	)
	createFlags.NetworkId = cmd.String(
		[]string{"-openstack-net-id"},
		"",
		"OpenStack network id the machine will be connected on",
	)
	createFlags.NetworkName = cmd.String(
		[]string{"-openstack-net-name"},
		"",
		"OpenStack network name the machine will be connected on",
	)
	createFlags.SecurityGroups = cmd.String(
		[]string{"-openstack-sec-groups"},
		"",
		"OpenStack comma separated security groups for the machine",
	)
	createFlags.FloatingIpPool = cmd.String(
		[]string{"-openstack-floatingip-pool"},
		"",
		"OpenStack floating IP pool to get an IP from to assign to the instance",
	)
	createFlags.SSHUser = cmd.String(
		[]string{"-openstack-ssh-user"},
		"root",
		"OpenStack SSH user. Set to root by default",
	)
	createFlags.SSHPort = cmd.Int(
		[]string{"-openstack-ssh-port"},
		22,
		"OpenStack SSH port. Set to 22 by default",
	)
	return createFlags
}

func NewDriver(storePath string) (drivers.Driver, error) {
	log.WithFields(log.Fields{
		"storePath": storePath,
	}).Debug("Instantiating OpenStack driver...")

	return NewDerivedDriver(storePath, &GenericClient{})
}

func NewDerivedDriver(storePath string, client Client) (*Driver, error) {
	return &Driver{
		storePath: storePath,
		client:    client,
	}, nil
}

func (d *Driver) DriverName() string {
	return "openstack"
}

func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
	flags := flagsInterface.(*CreateFlags)
	d.AuthUrl = *flags.AuthUrl
	d.Username = *flags.Username
	d.Password = *flags.Password
	d.TenantName = *flags.TenantName
	d.TenantId = *flags.TenantId
	d.Region = *flags.Region
	d.EndpointType = *flags.EndpointType
	d.FlavorName = *flags.FlavorName
	d.FlavorId = *flags.FlavorId
	d.ImageName = *flags.ImageName
	d.ImageId = *flags.ImageId
	d.NetworkId = *flags.NetworkId
	d.NetworkName = *flags.NetworkName
	if *flags.SecurityGroups != "" {
		d.SecurityGroups = strings.Split(*flags.SecurityGroups, ",")
	}
	d.FloatingIpPool = *flags.FloatingIpPool
	d.SSHUser = *flags.SSHUser
	d.SSHPort = *flags.SSHPort

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
	if err := d.initCompute(); err != nil {
		return "", err
	}

	addresses, err := d.client.GetInstanceIpAddresses(d)
	if err != nil {
		return "", err
	}

	floating := []string{}
	fixed := []string{}

	for _, address := range addresses {
		if address.AddressType == Floating {
			floating = append(floating, address.Address)
			continue
		}
		if address.AddressType == Fixed {
			fixed = append(fixed, address.Address)
			continue
		}
		log.Warnf("Unknown IP address type : %s", address)
	}

	if len(floating) == 1 {
		return d.foundIP(floating[0]), nil
	} else if len(floating) > 1 {
		log.Warnf("Multiple floating IP found. Take the first one of %s", floating)
		return d.foundIP(floating[0]), nil
	}

	if len(fixed) == 1 {
		return d.foundIP(fixed[0]), nil
	} else if len(fixed) > 1 {
		log.Warnf("Multiple fixed IP found. Take the first one of %s", floating)
		return d.foundIP(fixed[0]), nil
	}

	return "", fmt.Errorf("No IP found for the machine")
}

func (d *Driver) GetState() (state.State, error) {
	log.WithField("MachineId", d.MachineId).Debug("Get status for OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return state.None, err
	}
	if err := d.initNetwork(); err != nil {
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
	d.setMachineNameIfNotSet()
	d.KeyPairName = d.MachineName

	if err := d.resolveIds(); err != nil {
		return err
	}
	if err := d.createSSHKey(); err != nil {
		return err
	}
	if err := d.createMachine(); err != nil {
		return err
	}
	if err := d.waitForInstanceToStart(); err != nil {
		return err
	}
	if err := d.installDocker(); err != nil {
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
	return d.waitForInstanceToStart()
}

func (d *Driver) Stop() error {
	log.WithField("MachineId", d.MachineId).Info("Stopping OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.StopInstance(d); err != nil {
		return err
	}

	log.WithField("MachineId", d.MachineId).Info("Waiting for the OpenStack instance to stop...")
	if err := d.client.WaitForInstanceStatus(d, "SHUTOFF", 200); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Remove() error {
	log.WithField("MachineId", d.MachineId).Info("Deleting OpenStack instance...")
	if err := d.initCompute(); err != nil {
		return err
	}
	if err := d.client.DeleteInstance(d); err != nil {
		return err
	}
	log.WithField("Name", d.KeyPairName).Info("Deleting Key Pair...")
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
	return d.waitForInstanceToStart()
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("Upgrate is currently not available for the OpenStack driver")
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ip, err := d.GetIP()
	if err != nil {
		return nil, err
	}

	if d.SSHUser != "root" {
		cmd := strings.Replace(strings.Join(args, " "), "'", "\\'", -1)
		args = []string{"sudo", "sh", "-c", fmt.Sprintf("'%s'", cmd)}
	}

	log.Debug("Command: %s", args)
	return ssh.GetSSHCommand(ip, d.SSHPort, d.SSHUser, d.sshKeyPath(), args...), nil
}

const (
	errorMandatoryEnvOrOption    string = "%s must be specified either using the environment variable %s or the CLI option %s"
	errorMandatoryOption         string = "%s must be specified using the CLI option %s"
	errorExclusiveOptions        string = "Either %s or %s must be specified, not both"
	errorMandatoryTenantNameOrId string = "Tenant id or name must be provided either using one of the environment variables OS_TENANT_ID and OS_TENANT_NAME or one of the CLI options --openstack-tenant-id and --openstack-tenant-name"
	errorWrongEndpointType       string = "Endpoint type must be 'publicURL', 'adminURL' or 'internalURL'"
	errorUnknownFlavorName       string = "Unable to find flavor named %s"
	errorUnknownImageName        string = "Unable to find image named %s"
	errorUnknownNetworkName      string = "Unable to find network named %s"
)

func (d *Driver) checkConfig() error {
	if d.AuthUrl == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Autentication URL", "OS_AUTH_URL", "--openstack-auth-url")
	}
	if d.Username == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Username", "OS_USERNAME", "--openstack-username")
	}
	if d.Password == "" {
		return fmt.Errorf(errorMandatoryEnvOrOption, "Password", "OS_PASSWORD", "--openstack-password")
	}
	if d.TenantName == "" && d.TenantId == "" {
		return fmt.Errorf(errorMandatoryTenantNameOrId)
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

func (d *Driver) foundIP(ip string) string {
	log.WithFields(log.Fields{
		"IP":        ip,
		"MachineId": d.MachineId,
	}).Debug("IP address found")
	return ip
}

func (d *Driver) resolveIds() error {
	if d.NetworkName != "" {
		if err := d.initNetwork(); err != nil {
			return err
		}

		networkId, err := d.client.GetNetworkId(d, d.NetworkName)

		if err != nil {
			return err
		}

		if networkId == "" {
			return fmt.Errorf(errorUnknownNetworkName, d.NetworkName)
		}

		d.NetworkId = networkId
		log.WithFields(log.Fields{
			"Name": d.NetworkName,
			"ID":   d.NetworkId,
		}).Debug("Found network id using its name")
	}

	if d.FlavorName != "" {
		if err := d.initCompute(); err != nil {
			return err
		}
		flavorId, err := d.client.GetFlavorId(d, d.FlavorName)

		if err != nil {
			return err
		}

		if flavorId == "" {
			return fmt.Errorf(errorUnknownFlavorName, d.FlavorName)
		}

		d.FlavorId = flavorId
		log.WithFields(log.Fields{
			"Name": d.FlavorName,
			"ID":   d.FlavorId,
		}).Debug("Found flavor id using its name")
	}

	if d.ImageName != "" {
		if err := d.initCompute(); err != nil {
			return err
		}
		imageId, err := d.client.GetImageId(d, d.ImageName)

		if err != nil {
			return err
		}

		if imageId == "" {
			return fmt.Errorf(errorUnknownImageName, d.ImageName)
		}

		d.ImageId = imageId
		log.WithFields(log.Fields{
			"Name": d.ImageName,
			"ID":   d.ImageId,
		}).Debug("Found image id using its name")
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
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
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
	instanceId, err := d.client.CreateInstance(d)
	if err != nil {
		return err
	}
	d.MachineId = instanceId
	return nil
}

func (d *Driver) waitForInstanceToStart() error {
	log.WithField("MachineId", d.MachineId).Debug("Waiting for the OpenStack instance to start...")
	if err := d.client.WaitForInstanceStatus(d, "ACTIVE", 200); err != nil {
		return err
	}
	ip, err := d.GetIP()
	if err != nil {
		return err
	}
	return ssh.WaitForTCP(fmt.Sprintf("%s:%d", ip, d.SSHPort))
}

func (d *Driver) installDocker() error {
	log.WithField("MachineId", d.MachineId).Debug("Adding key to authorized-keys.d...")

	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/.docker/authorized-keys.d"); err != nil {
		return err
	}

	log.WithField("MachineId", d.MachineId).Debug("Installing docker daemon on the machine")

	cmd := "curl -sSL https://gist.githubusercontent.com/smashwilson/1a286139720a28ac6ead/raw/41d93c57ea2e86815cdfbfec42aaa696034afcc8/setup-docker.sh | /bin/bash"
	sshCmd, err := d.GetSSHCommand(cmd)
	if err != nil {
		return err
	}
	if err := sshCmd.Run(); err != nil {
		log.Warnf("Docker installation failed: %v", err)
		log.Warnf("The machine is not ready to run docker containers")
	}
	return nil
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-host-%s", utils.GenerateRandomID())
	}
}
