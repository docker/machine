package openstack

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type Driver struct {
	NameSuffix   string
	InstanceName string
	InstanceUUID string
	KeyName      string
	FlavorID     string
	ImageID      string
	IPAddress    string
	Region       string
	UserName     string
	DockerPort   string
	storePath    string
}

type CreateFlags struct {
	FlavorID   *string
	ImageID    *string
	Region     *string
	UserName   *string
	DockerPort *string
}

func init() {
	drivers.Register("openstack", &drivers.RegisteredDriver{
		New:                 NewDriver,
		RegisterCreateFlags: RegisterCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
	createFlags := new(CreateFlags)
	createFlags.FlavorID = cmd.String(
		[]string{"-openstack-flavor"},
		"1",
		"OpenStack flavor ID",
	)
	createFlags.ImageID = cmd.String(
		[]string{"-openstack-image"},
		"",
		"OpenStack image ID",
	)
	createFlags.Region = cmd.String(
		[]string{"-openstack-region"},
		"",
		"OpenStack region",
	)
	createFlags.UserName = cmd.String(
		[]string{"-openstack-username"},
		"root",
		"OpenStack instance username",
	)
	createFlags.DockerPort = cmd.String(
		[]string{"-openstack-docker-port"},
		"2375",
		"OpenStack docker port",
	)
	return createFlags
}

func NewDriver(storePath string) (drivers.Driver, error) {
	return &Driver{storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
	return "openstack"
}

func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
	flags := flagsInterface.(*CreateFlags)
	d.FlavorID = *flags.FlavorID
	d.ImageID = *flags.ImageID
	d.Region = *flags.Region
	d.UserName = *flags.UserName
	d.DockerPort = *flags.DockerPort

	if d.ImageID == "" {
		return fmt.Errorf("openstack driver requires the --openstack-image option")
	}

	return nil
}

func (d *Driver) Create() error {
	d.generateUniqueName()

	client, err := getClient()
	if err != nil {
		return err
	}

	log.Infof("Creating keypair...")
	if err := d.createKeypair(client); err != nil {
		return err
	}

	log.Infof("Creating server instance...")
	serverOpts := servers.CreateOpts{
		Name:      d.InstanceName,
		FlavorRef: d.FlavorID,
		ImageRef:  d.ImageID,
	}

	server, err := servers.Create(client, keypairs.CreateOptsExt{
		serverOpts,
		d.KeyName,
	}).Extract()
	if err != nil {
		return err
	}

	if err = servers.WaitForStatus(client, server.ID, "ACTIVE", 300); err != nil {
		log.Infof("Timeout waiting instance becoming ACTIVE")
		return err
	}
	d.InstanceUUID = server.ID

	details, err := servers.Get(client, d.InstanceUUID).Extract()
	for _, v := range details.Addresses {
		info := v.([]interface{})[0]
		dict := info.(map[string]interface{})
		d.IPAddress = dict["addr"].(string)
		log.Infof("Created instance name %s, ID %s, IP %s",
			d.InstanceName, d.InstanceUUID, d.IPAddress)
		// assuming instance with 1 network only
		break
	}

	if d.IPAddress == "" {
		return fmt.Errorf("No IP address found for Instance %s", d.InstanceUUID)
	}

	if err := d.waitForSSH(); err != nil {
		return err
	}

	if err := d.waitForDocker(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createKeypair(client *gophercloud.ServiceClient) error {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	pkBytes, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	key, err := keypairs.Create(client, keypairs.CreateOpts{
		Name:      fmt.Sprintf("docker-keypair-%s", d.NameSuffix),
		PublicKey: string(pkBytes),
	}).Extract()
	if err != nil {
		return err
	}

	d.KeyName = key.Name

	return err
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:%s", ip, d.DockerPort), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	client, err := getClient()
	if err != nil {
		return state.Error, err
	}

	details, err := servers.Get(client, d.InstanceUUID).Extract()
	switch details.Status {
	case "BUILDING":
		return state.Starting, nil
	case "ACTIVE":
		return state.Running, nil
	case "STOPPED":
		return state.Stopped, nil
	case "ERROR":
		return state.Error, nil
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	log.Infof("Start for OpenStack driver not supported yet...")
	return nil
}

func (d *Driver) Stop() error {
	log.Infof("Stop for OpenStack driver not supported yet...")
	return nil
}

func (d *Driver) Remove() error {
	client, err := getClient()
	if err != nil {
		return err
	}

	log.Infof("Deleting keypair...")
	if err = keypairs.Delete(client, d.KeyName).ExtractErr(); err != nil {
		return err
	}
	log.Infof("Deleting server instance...")
	if err = servers.Delete(client, d.InstanceUUID).ExtractErr(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	client, err := getClient()
	if err != nil {
		return err
	}

	log.Infof("Rebooting server instance name %s, ID %s...", d.InstanceName, d.InstanceUUID)
	res := servers.Reboot(client, d.InstanceUUID, servers.SoftReboot)
	if res.Err != nil {
		return res.Err
	}

	log.Infof("Waiting for instance turning ACTIVE...")
	if err = servers.WaitForStatus(client, d.InstanceUUID, "ACTIVE", 300); err != nil {
		log.Infof("Timeout waiting instance becoming ACTIVE")
		return err
	}

	if err := d.waitForSSH(); err != nil {
		return err
	}

	if err := d.waitForDocker(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	log.Infof("Kill for OpenStack driver not supported yet...")
	return nil
}

func (d *Driver) Upgrade() error {
	log.Infof("Upgrade for OpenStack driver not supported yet...")
	return nil
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, d.UserName, d.sshKeyPath(), args...), nil
}

func getClient() (*gophercloud.ServiceClient, error) {
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}

	return openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
}

func (d *Driver) waitForSSH() error {
	log.Infof("Waiting for SSH...")
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, 22)); err != nil {
		return err
	}
	return nil
}

func (d *Driver) waitForDocker() error {
	log.Infof("Waiting for docker daemon on host to be available...")
	url := fmt.Sprintf("%s:%d", d.IPAddress, 22)
	counter := 0
	retries := 12
	for {
		conn, err := net.Dial("tcp", url)
		if err != nil {
			time.Sleep(5 * time.Second)
			counter++
			if counter >= retries {
				return fmt.Errorf("timeout waiting docker daemon to be available")
			}
			continue
		}
		defer conn.Close()
		break
	}
	return nil
}

func (d *Driver) generateUniqueName() {
	id := utils.GenerateRandomID()[0:4]
	d.NameSuffix = id
	if d.InstanceName == "" {
		d.InstanceName = fmt.Sprintf("docker-host-%s", id)
	}
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}
