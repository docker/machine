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

const (
	dockerConfigDir = "/etc/docker"
)

type Driver struct {
	APIKey string

	MachineName    string
	IPAddress      string
	DockerPort     int
	CaCertPath     string
	PrivateKeyPath string

	client      *linodego.Client
	LinodeId    int
	LinodeLabel string

	DataCenterId   int
	PlanId         int
	PaymentTerm    int
	RootPassword   string
	SSHPort        int
	DistributionId int
	KernelId       int

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
		cli.IntFlag{
			EnvVar: "LINODE_SSH_PORT",
			Name:   "linode-ssh-port",
			Usage:  "Linode SSH Port",
			Value:  22,
		},
		cli.IntFlag{
			EnvVar: "LINODE_DISTRIBUTION_ID",
			Name:   "linode-distribution-id",
			Usage:  "Linode Distribution Id",
			Value:  124, // Ubuntu 14.04 LTS
		},
		cli.IntFlag{
			EnvVar: "LINODE_KERNEL_ID",
			Name:   "linode-kernel-id",
			Usage:  "Linode Kernel Id",
			Value:  138, // default kernel, Latest 64 bit (3.18.1-x86_64-linode50),
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey, DockerPort: 2376}, nil
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
	d.SSHPort = flags.Int("linode-ssh-port")
	d.DistributionId = flags.Int("linode-distribution-id")
	d.KernelId = flags.Int("linode-kernel-id")

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

func (d *Driver) PreCreateCheck() error {
	return nil
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
	sshCmd, err := d.GetSSHCommand("sudo apt-get update && sudo apt-get install --upgrade lxc-docker")
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

func (d *Driver) StartDocker() error {
	log.Debug("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker start")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) StopDocker() error {
	log.Debug("Stopping Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker stop")
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

// GetSSHCommand returns a command for SSH pointing at the correct user, host
// and keys for the host with args appended. If no args are passed, it will
// initiate an interactive SSH session as if SSH were passed no args.
func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, d.SSHPort, "root", d.sshKeyPath(), args...), nil
}

func (d *Driver) waitForJob(jobId int, jobName string) error {
	for {
		clientJobResponse, err := d.getClient().Job.List(d.LinodeId, jobId, false)
		if err != nil {
			return err
		}

		if len(clientJobResponse.Jobs) < 0 || clientJobResponse.Jobs[0].JobId != jobId {
			return errors.New(fmt.Sprintf("Job %s is not found.", jobName))
		}

		if clientJobResponse.Jobs[0].HostSuccess.String() == "1" {
			log.Debugf("Linode job %s completed.", jobName)
			return nil
		}

		log.Debugf("Wait for job %s completion...", jobName)
		time.Sleep(1 * time.Second)
	}
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
	distributionId := d.DistributionId
	createDiskJobResponse, err := d.client.Disk.CreateFromDistribution(distributionId, d.LinodeId, "Ubuntu Disk", 24576-256, args)

	if err != nil {
		return err
	}

	jobId := createDiskJobResponse.DiskJob.JobId
	diskId := createDiskJobResponse.DiskJob.DiskId
	log.Debugf("Linode create disk task :%d.", jobId)

	// wait until the creation is finished
	err = d.waitForJob(jobId, "Create Disk Task")
	if err != nil {
		return err
	}

	// create swap
	createDiskJobResponse, err = d.client.Disk.Create(d.LinodeId, "swap", "Swap Disk", 256, nil)
	if err != nil {
		return err
	}

	jobId = createDiskJobResponse.DiskJob.JobId
	swapDiskId := createDiskJobResponse.DiskJob.DiskId
	log.Debugf("Linode create swap disk task :%d.", jobId)

	// wait until the creation is finished
	err = d.waitForJob(jobId, "Create Swap Disk Task")
	if err != nil {
		return err
	}

	// create config
	args2 := make(map[string]string)
	args2["DiskList"] = fmt.Sprintf("%d,%d", diskId, swapDiskId)
	args2["RootDeviceNum"] = "1"
	args2["RootDeviceRO"] = "true"
	kernelId := d.KernelId
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
	jobId = jobResponse.JobId.JobId
	log.Debugf("Booting linode, job id: %v", jobId)
	// wait for boot
	err = d.waitForJob(jobId, "Booting linode")
	if err != nil {
		return err
	}

	log.Infof("Waiting for SSH...")

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, d.SSHPort)); err != nil {
		return err
	}

	// sleep for 10 seconds. This is required as Linode takes sometime to respond to SSH
	time.Sleep(10 * time.Second)

	log.Debugf("Setting hostname: %s", d.MachineName)
	cmd, err := d.GetSSHCommand(fmt.Sprintf(
		"echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts && sudo hostname %s && echo \"%s\" | sudo tee /etc/hostname",
		d.MachineName,
		d.MachineName,
		d.MachineName,
	))

	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Installing Docker")

	cmd, err = d.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl -sL https://get.docker.com | sh -; fi")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		log.Debugf("Error: cmd: %v Error: %v", cmd, err)
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
