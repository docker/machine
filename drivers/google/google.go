package google

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

// Driver is a struct compatible with the docker.hosts.drivers.Driver interface.
type Driver struct {
	*drivers.BaseDriver
	Zone          string
	MachineType   string
	MachineImage  string
	DiskType      string
	Address       string
	Preemptible   bool
	UseInternalIP bool
	Scopes        string
	DiskSize      int
	Project       string
	Tags          string
}

const (
	defaultZone        = "us-central1-a"
	defaultUser        = "docker-user"
	defaultMachineType = "n1-standard-1"
	defaultImageName   = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-1404-trusty-v20150909a"
	defaultScopes      = "https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write"
	defaultDiskType    = "pd-standard"
	defaultDiskSize    = 10
)

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "google-zone",
			Usage:  "GCE Zone",
			Value:  defaultZone,
			EnvVar: "GOOGLE_ZONE",
		},
		mcnflag.StringFlag{
			Name:   "google-machine-type",
			Usage:  "GCE Machine Type",
			Value:  defaultMachineType,
			EnvVar: "GOOGLE_MACHINE_TYPE",
		},
		mcnflag.StringFlag{
			Name:   "google-machine-image",
			Usage:  "GCE Machine Image Absolute URL",
			Value:  defaultImageName,
			EnvVar: "GOOGLE_MACHINE_IMAGE",
		},
		mcnflag.StringFlag{
			Name:   "google-username",
			Usage:  "GCE User Name",
			Value:  defaultUser,
			EnvVar: "GOOGLE_USERNAME",
		},
		mcnflag.StringFlag{
			Name:   "google-project",
			Usage:  "GCE Project",
			EnvVar: "GOOGLE_PROJECT",
		},
		mcnflag.StringFlag{
			Name:   "google-scopes",
			Usage:  "GCE Scopes (comma-separated if multiple scopes)",
			Value:  defaultScopes,
			EnvVar: "GOOGLE_SCOPES",
		},
		mcnflag.IntFlag{
			Name:   "google-disk-size",
			Usage:  "GCE Instance Disk Size (in GB)",
			Value:  defaultDiskSize,
			EnvVar: "GOOGLE_DISK_SIZE",
		},
		mcnflag.StringFlag{
			Name:   "google-disk-type",
			Usage:  "GCE Instance Disk type",
			Value:  defaultDiskType,
			EnvVar: "GOOGLE_DISK_TYPE",
		},
		mcnflag.StringFlag{
			Name:   "google-address",
			Usage:  "GCE Instance External IP",
			EnvVar: "GOOGLE_ADDRESS",
		},
		mcnflag.BoolFlag{
			Name:   "google-preemptible",
			Usage:  "GCE Instance Preemptibility",
			EnvVar: "GOOGLE_PREEMPTIBLE",
		},
		mcnflag.StringFlag{
			Name:   "google-tags",
			Usage:  "GCE Instance Tags (comma-separated)",
			EnvVar: "GOOGLE_TAGS",
			Value:  "",
		},
		mcnflag.BoolFlag{
			Name:   "google-use-internal-ip",
			Usage:  "Use internal GCE Instance IP rather than public one",
			EnvVar: "GOOGLE_USE_INTERNAL_IP",
		},
	}
}

// NewDriver creates a Driver with the specified storePath.
func NewDriver(machineName string, storePath string) *Driver {
	return &Driver{
		Zone:         defaultZone,
		DiskType:     defaultDiskType,
		DiskSize:     defaultDiskSize,
		MachineType:  defaultMachineType,
		MachineImage: defaultImageName,
		Scopes:       defaultScopes,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultUser,
			MachineName: machineName,
			StorePath:   storePath,
		},
	}
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetSSHUsername returns username for use with ssh
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
	d.Project = flags.String("google-project")
	if d.Project == "" {
		return fmt.Errorf("Please specify the Google Cloud Project name using the option --google-project.")
	}

	d.Zone = flags.String("google-zone")
	d.MachineType = flags.String("google-machine-type")
	d.MachineImage = flags.String("google-machine-image")
	d.DiskSize = flags.Int("google-disk-size")
	d.DiskType = flags.String("google-disk-type")
	d.Address = flags.String("google-address")
	d.Preemptible = flags.Bool("google-preemptible")
	d.UseInternalIP = flags.Bool("google-use-internal-ip")
	d.Scopes = flags.String("google-scopes")
	d.Tags = flags.String("google-tags")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = flags.String("google-username")
	d.SSHPort = 22

	return nil
}

// PreCreateCheck is called to enforce pre-creation steps
func (d *Driver) PreCreateCheck() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}

	// Check that the project exists. It will also check that credentials
	// at the same time.
	log.Infof("Check that the project exists")

	if _, err = c.service.Projects.Get(d.Project).Do(); err != nil {
		return fmt.Errorf("Project with ID %q not found. %v", d.Project, err)
	}

	// Check if the instance already exists. There will be an error if the instance
	// doesn't exist, so just check instance for nil.
	log.Infof("Check if the instance already exists")

	if instance, _ := c.instance(); instance != nil {
		return fmt.Errorf("Instance %v already exists.", d.MachineName)
	}

	return nil
}

// Create creates a GCE VM instance acting as a docker host.
func (d *Driver) Create() error {
	log.Infof("Generating SSH Key")

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating host...")

	c, err := newComputeUtil(d)
	if err != nil {
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

// Start starts an existing GCE instance or create an instance with an existing disk.
func (d *Driver) Start() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}

	instance, err := c.instance()
	if err != nil {
		if !strings.Contains(err.Error(), "notFound") {
			return err
		}
	}

	if instance == nil {
		if err = c.createInstance(d); err != nil {
			return err
		}
	} else {
		if err := c.startInstance(); err != nil {
			return err
		}
	}

	d.IPAddress, err = d.GetIP()
	return err
}

// Stop stops an existing GCE instance.
func (d *Driver) Stop() error {
	c, err := newComputeUtil(d)
	if err != nil {
		return err
	}

	if err := c.stopInstance(); err != nil {
		return err
	}

	d.IPAddress = ""
	return nil
}

// Kill stops an existing GCE instance.
func (d *Driver) Kill() error {
	return d.Stop()
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

func (d *Driver) Restart() error {
	return nil
}
