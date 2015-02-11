package softlayer

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	dockerConfigDir  = "/etc/docker"
	ApiEndpoint      = "https://api.softlayer.com/rest/v3"
	DockerInstallUrl = "https://get.docker.com"
)

type Driver struct {
	storePath      string
	IPAddress      string
	deviceConfig   *deviceConfig
	Id             int
	Client         *Client
	MachineName    string
	CaCertPath     string
	PrivateKeyPath string
}

type deviceConfig struct {
	DiskSize      int
	Cpu           int
	Hostname      string
	Domain        string
	Region        string
	Memory        int
	Image         string
	HourlyBilling bool
	InstallScript string
	LocalDisk     bool
	PrivateNet    bool
}

func init() {
	drivers.Register("softlayer", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func GetCreateFlags() []cli.Flag {
	// Set hourly billing to true by default since codegangsta cli doesn't take default bool values
	if os.Getenv("SOFTLAYER_HOURLY_BILLING") == "" {
		os.Setenv("SOFTLAYER_HOURLY_BILLING", "true")
	}
	return []cli.Flag{
		cli.IntFlag{
			EnvVar: "SOFTLAYER_MEMORY",
			Name:   "softlayer-memory",
			Usage:  "Memory in MB for machine",
			Value:  1024,
		},
		cli.IntFlag{
			EnvVar: "SOFTLAYER_DISK_SIZE",
			Name:   "softlayer-disk-size",
			Usage:  "Disk size for machine, a value of 0 uses the default size on softlayer",
			Value:  0,
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_USER",
			Name:   "softlayer-user",
			Usage:  "softlayer user account name",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_API_KEY",
			Name:   "softlayer-api-key",
			Usage:  "softlayer user API key",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_REGION",
			Name:   "softlayer-region",
			Usage:  "softlayer region for machine",
			Value:  "dal01",
		},
		cli.IntFlag{
			EnvVar: "SOFTLAYER_CPU",
			Name:   "softlayer-cpu",
			Usage:  "number of CPU's for the machine",
			Value:  1,
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_HOSTNAME",
			Name:   "softlayer-hostname",
			Usage:  "hostname for the machine",
			Value:  "docker",
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_DOMAIN",
			Name:   "softlayer-domain",
			Usage:  "domain name for machine",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_API_ENDPOINT",
			Name:   "softlayer-api-endpoint",
			Usage:  "softlayer api endpoint to use",
			Value:  ApiEndpoint,
		},
		cli.BoolFlag{
			EnvVar: "SOFTLAYER_HOURLY_BILLING",
			Name:   "softlayer-hourly-billing",
			Usage:  "set hourly billing for machine - on by default",
		},
		cli.BoolFlag{
			EnvVar: "SOFTLAYER_LOCAL_DISK",
			Name:   "softlayer-local-disk",
			Usage:  "use machine local disk instead of softlayer SAN",
		},
		cli.BoolFlag{
			EnvVar: "SOFTLAYER_PRIVATE_NET",
			Name:   "softlayer-private-net-only",
			Usage:  "Use only private networking",
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_IMAGE",
			Name:   "softlayer-image",
			Usage:  "OS image for machine",
			Value:  "UBUNTU_LATEST",
		},
		cli.StringFlag{
			EnvVar: "SOFTLAYER_INSTALL_SCRIPT",
			Name:   "softlayer-install-script",
			Usage:  "Install script to call after the machine is initialized (should install Docker)",
			Value:  DockerInstallUrl,
		},
	}
}

func validateDeviceConfig(c *deviceConfig) error {
	if c.Hostname == "" {
		return fmt.Errorf("Missing required setting - --softlayer-hostname")
	}
	if c.Domain == "" {
		return fmt.Errorf("Missing required setting - --softlayer-domain")
	}

	if c.Region == "" {
		return fmt.Errorf("Missing required setting - --softlayer-region")
	}
	if c.Cpu < 1 {
		return fmt.Errorf("Missing required setting - --softlayer-cpu")
	}

	return nil
}

func validateClientConfig(c *Client) error {
	if c.ApiKey == "" {
		return fmt.Errorf("Missing required setting - --softlayer-api-key")
	}

	if c.User == "" {
		return fmt.Errorf("Missing required setting - --softlayer-user")
	}

	if c.Endpoint == "" {
		return fmt.Errorf("Missing required setting - --softlayer-api-endpoint")
	}

	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {

	d.Client = &Client{
		Endpoint: flags.String("softlayer-api-endpoint"),
		User:     flags.String("softlayer-user"),
		ApiKey:   flags.String("softlayer-api-key"),
	}

	if err := validateClientConfig(d.Client); err != nil {
		return err
	}

	d.deviceConfig = &deviceConfig{
		Hostname:      flags.String("softlayer-hostname"),
		DiskSize:      flags.Int("softlayer-disk-size"),
		Cpu:           flags.Int("softlayer-cpu"),
		Domain:        flags.String("softlayer-domain"),
		Memory:        flags.Int("softlayer-memory"),
		PrivateNet:    flags.Bool("softlayer-private-net-only"),
		LocalDisk:     flags.Bool("softlayer-local-disk"),
		HourlyBilling: flags.Bool("softlayer-hourly-billing"),
		InstallScript: flags.String("softlayer-install-script"),
		Image:         "UBUNTU_LATEST",
		Region:        flags.String("softlayer-region"),
	}
	return validateDeviceConfig(d.deviceConfig)
}

func (d *Driver) getClient() *Client {
	return d.Client
}

func (d *Driver) DriverName() string {
	return "softlayer"
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

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return "tcp://" + ip + ":2376", nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress != "" {
		return d.IPAddress, nil
	}
	return d.getClient().VirtualGuest().GetPublicIp(d.Id)
}

func (d *Driver) GetState() (state.State, error) {
	s, err := d.getClient().VirtualGuest().PowerState(d.Id)
	if err != nil {
		return state.None, err
	}
	var vmState state.State
	switch s {
	case "Running":
		vmState = state.Running
	case "Halted":
		vmState = state.Stopped
	default:
		vmState = state.None
	}
	return vmState, nil
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, "root", d.sshKeyPath(), args...), nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	waitForStart := func() {
		log.Infof("Waiting for host to become available")
		for {
			s, err := d.GetState()
			if err != nil {
				continue
			}

			if s == state.Running {
				break
			}
			time.Sleep(2 * time.Second)
		}
	}

	getIp := func() {
		log.Infof("Getting Host IP")
		for {
			var (
				ip  string
				err error
			)
			if d.deviceConfig.PrivateNet {
				ip, err = d.getClient().VirtualGuest().GetPrivateIp(d.Id)
			} else {
				ip, err = d.getClient().VirtualGuest().GetPublicIp(d.Id)
			}
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			// not a perfect regex, but should be just fine for our needs
			exp := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
			if exp.MatchString(ip) {
				d.IPAddress = ip
				break
			}
			time.Sleep(2 * time.Second)
		}
	}

	log.Infof("Creating SSH key...")
	key, err := d.createSSHKey()
	if err != nil {
		return err
	}

	spec := d.buildHostSpec()
	spec.SshKeys = []*SshKey{key}

	id, err := d.getClient().VirtualGuest().Create(spec)
	if err != nil {
		return fmt.Errorf("Error creating host: %q", err)
	}
	d.Id = id
	getIp()
	waitForStart()
	ssh.WaitForTCP(d.IPAddress + ":22")
	if err := d.setupHost(); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up host config: %q", err)
	}
	return nil
}

func (d *Driver) buildHostSpec() *HostSpec {
	spec := &HostSpec{
		Hostname:       d.deviceConfig.Hostname,
		Domain:         d.deviceConfig.Domain,
		Cpu:            d.deviceConfig.Cpu,
		Memory:         d.deviceConfig.Memory,
		Datacenter:     Datacenter{Name: d.deviceConfig.Region},
		InstallScript:  d.deviceConfig.InstallScript,
		Os:             d.deviceConfig.Image,
		HourlyBilling:  d.deviceConfig.HourlyBilling,
		PrivateNetOnly: d.deviceConfig.PrivateNet,
	}
	if d.deviceConfig.DiskSize > 0 {
		spec.BlockDevices = []BlockDevice{{Device: "0", DiskImage: DiskImage{Capacity: d.deviceConfig.DiskSize}}}
	}
	return spec
}

func (d *Driver) createSSHKey() (*SshKey, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	key, err := d.getClient().SshKey().Create(d.deviceConfig.Hostname, string(publicKey))
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) Kill() error {
	return d.getClient().VirtualGuest().PowerOff(d.Id)
}
func (d *Driver) Remove() error {
	var err error
	for i := 0; i < 5; i++ {
		if err = d.getClient().VirtualGuest().Cancel(d.Id); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	return err
}
func (d *Driver) Restart() error {
	return d.getClient().VirtualGuest().Reboot(d.Id)
}
func (d *Driver) Start() error {
	return d.getClient().VirtualGuest().PowerOn(d.Id)
}
func (d *Driver) Stop() error {
	return d.getClient().VirtualGuest().PowerOff(d.Id)
}

func (d *Driver) Upgrade() error {
	log.Debugf("Upgrading Docker")

	cmd, err := d.GetSSHCommand("sudo apt-get update && sudo apt-get install --upgrade lxc-docker")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
		return err

	}

	return cmd.Run()
}

func (d *Driver) setupHost() error {
	log.Infof("Configuring host OS")
	ssh.WaitForTCP(d.IPAddress + ":22")
	// Wait to make sure docker is installed
	for {
		cmd, err := d.GetSSHCommand(`[ -f "$(which docker)" ] && [ -f "/etc/default/docker" ] || exit 1`)
		if err != nil {
			return err
		}
		if err := cmd.Run(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	return nil
}
