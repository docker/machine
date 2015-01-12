package linode

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/taoh/linodego"
)

type Driver struct {
	APIKey string

	IPAddress  string
	DockerPort int

	client      *linodego.Client
	LinodeId    int
	LinodeLabel string

	DataCenterId int
	PlanId       int
	PaymentTerm  int
	RootPassword string

	storePath string
}

func init() {
	drivers.Register("linode", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "LINODE_API_KEY",
			Name:   "linode-api-key",
			Usage:  "Linode API Key",
		},
		cli.StringFlag{
			EnvVar: "LINODE_ROOT_PASSWORD",
			Name:   "linode-root-pass",
			Usage:  "Root password",
		},
		cli.IntFlag{
			EnvVar: "LINODE_DATACENTER_ID",
			Name:   "linode-datacenter-id",
			Usage:  "Linode Data Center Id",
			Value:  2,
		},
		cli.IntFlag{
			EnvVar: "LINODE_PLAN_ID",
			Name:   "linode-plan-id",
			Usage:  "Linode plan id",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "LINODE_PAYMENT_TERM",
			Name:   "linode-payment-term",
			Usage:  "Linode Payment term",
			Value:  1, // valid values: 1, 12, 24
		},
	}
}

func NewDriver(storePath string) (drivers.Driver, error) {
	return &Driver{storePath: storePath, DockerPort: 2376}, nil
}

func (d *Driver) DriverName() string {
	return "linode"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIKey = flags.String("linode-api-key")
	d.DataCenterId = flags.Int("linode-datacenter-id")
	d.PlanId = flags.Int("linode-plan-id")
	d.PaymentTerm = flags.Int("linode-payment-term")
	d.RootPassword = flags.String("linode-root-pass")

	if d.APIKey == "" {
		return fmt.Errorf("linode driver requires the --linode-api-key option")
	}

	if d.RootPassword == "" {
		return fmt.Errorf("linode driver requires the --linode-root-pass option")
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:%d", ip, d.DockerPort), nil
}

// Get IP Address for the Linode. Note that currently the IP Address
// is cached
func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

// Get Linode Client
func (d *Driver) getClient() *linodego.Client {
	if d.client == nil {
		d.client = linodego.NewClient(d.APIKey, nil)
	}
	return d.client
}

// Get State
func (d *Driver) GetState() (state.State, error) {
	linodes, err := d.getClient().Linode.List(d.LinodeId)
	if err != nil {
		return state.Error, err
	}

	// Status flag values:
	// -2: Boot Failed (not in use)
	// -1: Being Created
	//  0: Brand New
	//  1: Running
	//  2: Powered Off
	//  3: Shutting Down (not in use)
	//  4: Saved to Disk (not in use)
	//
	switch linodes.Linodes[0].Status {
	case 0:
		return state.Starting, nil
	case 1:
		return state.Running, nil
	case 2:
		return state.Stopped, nil
	case 3:
		return state.Stopping, nil
	}
	return state.None, nil
}

// Start a host
func (d *Driver) Start() error {
	_, err := d.getClient().Linode.Boot(d.LinodeId, -1)
	return err
}

// Stop a host
func (d *Driver) Stop() error {
	_, err := d.getClient().Linode.Shutdown(d.LinodeId)
	return err
}

// Restart a host.
func (d *Driver) Restart() error {
	_, err := d.getClient().Linode.Reboot(d.LinodeId, -1)
	return err
}

// Kill stops a host forcefully, for Linode it is the same as Stop
func (d *Driver) Kill() error {
	_, err := d.getClient().Linode.Shutdown(d.LinodeId)
	return err
}

// Remove the host
func (d *Driver) Remove() error {
	client := d.getClient()
	log.Debugf("Removing linode: %d", d.LinodeId)
	if _, err := client.Linode.Delete(d.LinodeId, true); err != nil {
		return err
	}
	return nil
}

// Upgrade the version of Docker on the host to the latest version
func (d *Driver) Upgrade() error {
	sshCmd, err := d.GetSSHCommand("apt-get update && apt-get install lxc-docker")
	if err != nil {
		return err
	}
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

// GetSSHCommand returns a command for SSH pointing at the correct user, host
// and keys for the host with args appended. If no args are passed, it will
// initiate an interactive SSH session as if SSH were passed no args.
func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, "root", d.sshKeyPath(), args...), nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	publicKey, err := d.createSSHKey()
	if err != nil {
		return err
	}

	log.Infof("Creating linode...")

	client := d.getClient()

	// Create a linode
	linodeResponse, err := client.Linode.Create(
		d.DataCenterId,
		d.PlanId,
		d.PaymentTerm,
	)
	if err != nil {
		return err
	}

	d.LinodeId = linodeResponse.LinodeId.LinodeId
	log.Debugf("Linode created: %d", d.LinodeId)

	linodeIPListResponse, err := client.Ip.List(d.LinodeId, -1)
	if err != nil {
		return err
	}
	for _, fullIpAddress := range linodeIPListResponse.FullIPAddresses {
		if fullIpAddress.IsPublic == 1 {
			d.IPAddress = fullIpAddress.IPAddress
		}
	}

	if d.IPAddress == "" {
		return errors.New("Linode IP Address is not found.")
	}

	log.Debugf("Created linode ID %d, IP address %s",
		d.LinodeId,
		d.IPAddress)

	// Deploy distribution
	args := make(map[string]string)
	args["rootPass"] = d.RootPassword
	args["rootSSHKey"] = publicKey
	distributionId := 124
	createDiskJobResponse, err := d.client.Disk.CreateFromDistribution(distributionId, d.LinodeId, "Ubuntu Disk", 24576-256, args)

	if err != nil {
		return err
	}

	jobId := createDiskJobResponse.DiskJob.JobId
	diskId := createDiskJobResponse.DiskJob.DiskId
	log.Debugf("Linode create disk task :%d.", jobId)

	// wait until the creation is finished
	for {
		clientJobResponse, err := client.Job.List(d.LinodeId, jobId, false)
		if err != nil {
			return err
		}

		if len(clientJobResponse.Jobs) < 0 || clientJobResponse.Jobs[0].JobId != jobId {
			return errors.New("Create Disk Job is not created.")
		}

		if clientJobResponse.Jobs[0].HostSuccess.String() == "1" {
			log.Debugf("Linode create disk task completed.")
			break
		}

		log.Debugf("Wait for creating disk task")
		time.Sleep(1 * time.Second)
	}

	// create swap
	//
	createDiskJobResponse, err = d.client.Disk.Create(d.LinodeId, "swap", "Swap Disk", 256, nil)
	if err != nil {
		return err
	}

	jobId = createDiskJobResponse.DiskJob.JobId
	swapDiskId := createDiskJobResponse.DiskJob.DiskId
	log.Debugf("Linode create swap disk task :%d.", jobId)

	// wait until the creation is finished
	for {
		clientJobResponse, err := client.Job.List(d.LinodeId, jobId, false)
		if err != nil {
			return err
		}

		if len(clientJobResponse.Jobs) < 0 || clientJobResponse.Jobs[0].JobId != jobId {
			return errors.New("Create Disk Job is not created.")
		}

		if clientJobResponse.Jobs[0].HostSuccess.String() == "1" {
			log.Debugf("Linode create swap disk task completed.")
			break
		}

		log.Debugf("Wait for creating swap disk task")
		time.Sleep(1 * time.Second)
	}

	// create config
	args2 := make(map[string]string)
	args2["DiskList"] = fmt.Sprintf("%d,%d", diskId, swapDiskId)
	args2["RootDeviceNum"] = "1"
	args2["RootDeviceRO"] = "true"
	kernelId := 138 // default kernel, Latest 64 bit (3.18.1-x86_64-linode50)
	_, err = d.client.Config.Create(d.LinodeId, kernelId, "My Machine Configuration", args2)

	if err != nil {
		return err
	}

	log.Debugf("Linode configuration created :%d.", jobId)

	// Boot
	jobResponse, err := d.client.Linode.Boot(d.LinodeId, -1)
	if err != nil {
		return err
	}
	log.Debugf("Booting linode, job id: %v", jobResponse.JobId.JobId)
	// wait for boot
	for {
		clientJobResponse, err := client.Job.List(d.LinodeId, jobResponse.JobId.JobId, false)
		if err != nil {
			return err
		}

		if len(clientJobResponse.Jobs) < 0 {
			return errors.New("Boot Job is not created.")
		}

		hostSuccess := clientJobResponse.Jobs[0].HostSuccess.String()
		if hostSuccess == "1" {
			log.Debugf("Linode wait for Boot completed.")
			break
		}

		if hostSuccess == "0" {
			return errors.New("Failed to boot linode")
		}

		log.Debugf("Wait for boot task")
		time.Sleep(1 * time.Second)
	}

	log.Infof("Waiting for SSH...")

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, 22)); err != nil {
		return err
	}

	// sleep for 10 seconds. This is required as Linode takes sometime to respond to SSH
	time.Sleep(10 * time.Second)

	log.Debugf("Installing docker ...")
	cmd, err := d.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl get.docker.io | sudo sh -; fi")
	if err := cmd.Run(); err != nil {
		log.Debugf("Error: cmd: %v Error: %v", cmd, err)
		return err
	}

	log.Debugf("Stopping docker...")
	cmd, err = d.GetSSHCommand("sudo stop docker")
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("HACK: Downloading version of Docker with identity auth...")

	cmd, err = d.GetSSHCommand("curl -sS https://ehazlett.s3.amazonaws.com/public/docker/linux/docker-1.4.1-136b351e-identity > /usr/bin/docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Updating /etc/default/docker to use identity auth...")

	cmd, err = d.GetSSHCommand(fmt.Sprintf("echo 'export DOCKER_OPTS=\"--auth=identity --host=tcp://0.0.0.0:%d\"' >> /etc/default/docker", d.DockerPort))
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Adding key to authorized-keys.d...")

	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/.docker/authorized-keys.d"); err != nil {
		return err
	}

	log.Debugf("Starting docker...")
	cmd, err = d.GetSSHCommand("sudo start docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) createSSHKey() (string, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}
