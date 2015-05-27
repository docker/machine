package google

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

// Driver is a struct compatible with the docker.hosts.drivers.Driver interface.
type Driver struct {
	*drivers.BaseDriver
	Zone          string
	MachineType   string
	DiskType      string
	Scopes        string
	DiskSize      int
	AuthTokenPath string
	Project       string
}

func init() {
	drivers.Register("google", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "google-zone",
			Usage:  "GCE Zone",
			Value:  "us-central1-a",
			EnvVar: "GOOGLE_ZONE",
		},
		cli.StringFlag{
			Name:   "google-machine-type",
			Usage:  "GCE Machine Type",
			Value:  "f1-micro",
			EnvVar: "GOOGLE_MACHINE_TYPE",
		},
		cli.StringFlag{
			Name:   "google-username",
			Usage:  "GCE User Name",
			Value:  "docker-user",
			EnvVar: "GOOGLE_USERNAME",
		},
		cli.StringFlag{
			Name:   "google-project",
			Usage:  "GCE Project",
			EnvVar: "GOOGLE_PROJECT",
		},
		cli.StringFlag{
			Name:   "google-auth-token",
			Usage:  "GCE oAuth token",
			EnvVar: "GOOGLE_AUTH_TOKEN",
		},
		cli.StringFlag{
			Name:   "google-scopes",
			Usage:  "GCE Scopes (comma-separated if multiple scopes)",
			Value:  "https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write",
			EnvVar: "GOOGLE_SCOPES",
		},
		cli.IntFlag{
			Name:   "google-disk-size",
			Usage:  "GCE Instance Disk Size (in GB)",
			Value:  10,
			EnvVar: "GOOGLE_DISK_SIZE",
		},
		cli.StringFlag{
			Name:   "google-disk-type",
			Usage:  "GCE Instance Disk type",
			Value:  "pd-standard",
			EnvVar: "GOOGLE_DISK_TYPE",
		},
	}
}

// NewDriver creates a Driver with the specified storePath.
func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	inner := drivers.NewBaseDriver(machineName, storePath, caCert, privateKey)
	return &Driver{BaseDriver: inner}, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker-user"
	}

	return d.SSHUser
}

// DriverName returns the name of the driver.
func (d *Driver) DriverName() string {
	return "google"
}

// SetConfigFromFlags initializes the driver based on the command line flags.
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Zone = flags.String("google-zone")
	d.MachineType = flags.String("google-machine-type")
	d.DiskSize = flags.Int("google-disk-size")
	d.DiskType = flags.String("google-disk-type")
	d.AuthTokenPath = flags.String("google-auth-token")
	d.Project = flags.String("google-project")
	d.Scopes = flags.String("google-scopes")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	if d.Project == "" {
		return fmt.Errorf("Please specify the Google Cloud Project name using the option --google-project.")
	}
	d.SSHUser = flags.String("google-username")
	d.SSHPort = 22
	return nil
}

func (d *Driver) initApis() (*ComputeUtil, error) {
	return newComputeUtil(d)
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

// Create creates a GCE VM instance acting as a docker host.
func (d *Driver) Create() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}
	log.Infof("Creating host...")
	// Check if the instance already exists. There will be an error if the instance
	// doesn't exist, so just check instance for nil.
	if instance, _ := c.instance(); instance != nil {
		return fmt.Errorf("Instance %v already exists.", d.MachineName)
	}

	log.Infof("Generating SSH Key")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	return c.createInstance(d)
}

// GetURL returns the URL of the remote docker daemon.
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("tcp://%s:2376", ip)
	return url, nil
}

// GetIP returns the IP address of the GCE instance.
func (d *Driver) GetIP() (string, error) {
	c, err := newComputeUtil(d)
	if err != nil {
		return "", err
	}
	return c.ip()
}

// GetState returns a docker.hosts.state.State value representing the current state of the host.
func (d *Driver) GetState() (state.State, error) {
	c, err := newComputeUtil(d)
	if err != nil {
		return state.None, err
	}

	// All we care about is whether the disk exists, so we just check disk for a nil value.
	// There will be no error if disk is not nil.
	disk, _ := c.disk()
	instance, _ := c.instance()
	if instance == nil && disk == nil {
		return state.None, nil
	}
	if instance == nil && disk != nil {
		return state.Stopped, nil
	}

	switch instance.Status {
	case "PROVISIONING", "STAGING":
		return state.Starting, nil
	case "RUNNING":
		return state.Running, nil
	case "STOPPING", "STOPPED", "TERMINATED":
		return state.Stopped, nil
	}
	return state.None, nil
}

// Start creates a GCE instance and attaches it to the existing disk.
func (d *Driver) Start() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}
	if err = c.createInstance(d); err != nil {
		return err
	}
	d.IPAddress, err = d.GetIP()
	return err
}

// Stop deletes the GCE instance, but keeps the disk.
func (d *Driver) Stop() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}
	if err = c.deleteInstance(); err != nil {
		return err
	}
	d.IPAddress = ""
	return nil
}

// Remove deletes the GCE instance and the disk.
func (d *Driver) Remove() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}
	s, err := d.GetState()
	if err != nil {
		return err
	}
	if s == state.Running {
		if err := c.deleteInstance(); err != nil {
			return err
		}
	}
	return c.deleteDisk()
}

// Restart deletes and recreates the GCE instance, keeping the disk.
func (d *Driver) Restart() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}
	if err := c.deleteInstance(); err != nil {
		return err
	}

	return c.createInstance(d)
}

// Kill deletes the GCE instance, but keeps the disk.
func (d *Driver) Kill() error {
	return d.Stop()
}
