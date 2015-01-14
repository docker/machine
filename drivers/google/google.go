package google

import (
	"fmt"

	"github.com/docker/machine/state"

	"os/exec"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
)

// Driver is a struct compatible with the docker.hosts.drivers.Driver interface.
type Driver struct {
	MachineName      string
	Zone             string
	MachineType      string
	storePath        string
	UserName         string
	Project          string
	sshKeyPath       string
	publicSSHKeyPath string
}

// CreateFlags are the command line flags used to create a driver.
type CreateFlags struct {
	Zone        *string
	MachineType *string
	UserName    *string
	Project     *string
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
			Usage:  "User Name",
			Value:  "docker-user",
			EnvVar: "GOOGLE_USERNAME",
		},
		cli.StringFlag{
			Name:   "google-project",
			Usage:  "GCE Project",
			EnvVar: "GOOGLE_PROJECT",
		},
	}
}

// NewDriver creates a Driver with the specified storePath.
func NewDriver(machineName string, storePath string) (drivers.Driver, error) {
	return &Driver{
		MachineName:      machineName,
		storePath:        storePath,
		sshKeyPath:       path.Join(storePath, "id_rsa"),
		publicSSHKeyPath: path.Join(storePath, "id_rsa.pub"),
	}, nil
}

// DriverName returns the name of the driver.
func (driver *Driver) DriverName() string {
	return "google"
}

// SetConfigFromFlags initializes the driver based on the command line flags.
func (driver *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	driver.Zone = flags.String("google-zone")
	driver.MachineType = flags.String("google-machine-type")
	driver.UserName = flags.String("google-username")
	driver.Project = flags.String("google-project")
	if driver.Project == "" {
		return fmt.Errorf("Please specify the Google Cloud Project name using the option --google-project.")
	}
	return nil
}

func (driver *Driver) initApis() (*ComputeUtil, error) {
	return newComputeUtil(driver)
}

// Create creates a GCE VM instance acting as a docker host.
func (driver *Driver) Create() error {
	c, err := newComputeUtil(driver)
	if err != nil {
		return err
	}
	log.Infof("Creating host...")
	// Check if the instance already exists. There will be an error if the instance
	// doesn't exist, so just check instance for nil.
	if instance, _ := c.instance(); instance != nil {
		return fmt.Errorf("Instance %v already exists.", driver.MachineName)
	}

	log.Infof("Generating SSH Key")
	if err := ssh.GenerateSSHKey(driver.sshKeyPath); err != nil {
		return err
	}

	return c.createInstance(driver)
}

// GetURL returns the URL of the remote docker daemon.
func (driver *Driver) GetURL() (string, error) {
	ip, err := driver.GetIP()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("tcp://%s:2376", ip)
	return url, nil
}

// GetIP returns the IP address of the GCE instance.
func (driver *Driver) GetIP() (string, error) {
	c, err := newComputeUtil(driver)
	if err != nil {
		return "", err
	}
	return c.ip()
}

// GetState returns a docker.hosts.state.State value representing the current state of the host.
func (driver *Driver) GetState() (state.State, error) {
	c, err := newComputeUtil(driver)
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
func (driver *Driver) Start() error {
	c, err := newComputeUtil(driver)
	if err != nil {
		return err
	}
	return c.createInstance(driver)
}

// Stop deletes the GCE instance, but keeps the disk.
func (driver *Driver) Stop() error {
	c, err := newComputeUtil(driver)
	if err != nil {
		return err
	}
	return c.deleteInstance()
}

// Remove deletes the GCE instance and the disk.
func (driver *Driver) Remove() error {
	c, err := newComputeUtil(driver)
	if err != nil {
		return err
	}
	s, err := driver.GetState()
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
func (driver *Driver) Restart() error {
	c, err := newComputeUtil(driver)
	if err != nil {
		return err
	}
	if err := c.deleteInstance(); err != nil {
		return err
	}

	return c.createInstance(driver)
}

// Kill deletes the GCE instance, but keeps the disk.
func (driver *Driver) Kill() error {
	return driver.Stop()
}

// GetSSHCommand returns a command that will run over SSH on the GCE instance.
func (driver *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ip, err := driver.GetIP()
	if err != nil {
		return nil, err
	}
	return ssh.GetSSHCommand(ip, 22, driver.UserName, driver.sshKeyPath, args...), nil
}

// Upgrade upgrades the docker daemon on the host to the latest version.
func (driver *Driver) Upgrade() error {
	c, err := newComputeUtil(driver)
	if err != nil {
		return err
	}
	return c.updateDocker(driver)
}
