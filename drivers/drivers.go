package drivers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/state"
)

// Driver defines how a host is created and controlled. Different types of
// driver represent different ways hosts can be created (e.g. different
// hypervisors, different cloud providers)
type Driver interface {
	// DriverName returns the name of the driver as it is registered
	DriverName() string

	// SetConfigFromFlags configures the driver with the object that was returned
	// by RegisterCreateFlags
	SetConfigFromFlags(flags DriverOptions) error

	// GetURL returns a Docker compatible host URL for connecting to this host
	// e.g. tcp://1.2.3.4:2376
	GetURL() (string, error)

	// GetIP returns an IP or hostname that this host is available at
	// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
	GetIP() (string, error)

	// GetSSHUser returns the user for ssh
	GetSSHUser() string

	// GetSSHPort returns the port for ssh
	GetSSHPort() int

	// GetMachineOptions returns the machine options for the driver
	GetMachineOptions() (*MachineOptions, error)

	// GetState returns the state that the host is in (running, stopped, etc)
	GetState() (state.State, error)

	// Create a host using the driver's config
	Create() error

	// Remove a host
	Remove() error

	// Start a host
	Start() error

	// Stop a host gracefully
	Stop() error

	// Restart a host. This may just call Stop(); Start() if the provider does not
	// have any special restart behaviour.
	Restart() error

	// Kill stops a host forcefully
	Kill() error

	// Upgrade the version of Docker on the host to the latest version
	Upgrade() error

	// GetSSHCommand returns a command for SSH pointing at the correct user, host
	// and keys for the host with args appended. If no args are passed, it will
	// initiate an interactive SSH session as if SSH were passed no args.
	GetSSHCommand(args ...string) (*exec.Cmd, error)
}

type CloudConfigOptions struct {
	PublicKey  string
	DockerOpts string
}

// RegisteredDriver is used to register a driver with the Register function.
// It has two attributes:
// - New: a function that returns a new driver given a path to store host
//   configuration in
// - RegisterCreateFlags: a function that takes the FlagSet for
//   "docker hosts create" and returns an object to pass to SetConfigFromFlags
type RegisteredDriver struct {
	New            func(storePath string) (Driver, error)
	GetCreateFlags func() []cli.Flag
}

var ErrHostIsNotRunning = errors.New("host is not running")

var (
	drivers map[string]*RegisteredDriver
)

func init() {
	drivers = make(map[string]*RegisteredDriver)
}

// Register a driver
func Register(name string, registeredDriver *RegisteredDriver) error {
	if _, exists := drivers[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
	}

	drivers[name] = registeredDriver
	return nil
}

// NewDriver creates a new driver of type "name"
func NewDriver(name string, storePath string) (Driver, error) {
	driver, exists := drivers[name]
	if !exists {
		return nil, fmt.Errorf("hosts: Unknown driver %q", name)
	}
	return driver.New(storePath)
}

// GetCreateFlags runs GetCreateFlags for all of the drivers and
// returns their return values indexed by the driver name
func GetCreateFlags() []cli.Flag {
	flags := []cli.Flag{}

	for driverName := range drivers {
		driver := drivers[driverName]
		for _, f := range driver.GetCreateFlags() {
			flags = append(flags, f)
		}
	}

	sort.Sort(ByFlagName(flags))

	return flags
}

// GetDriverNames returns a slice of all registered driver names
func GetDriverNames() []string {
	names := make([]string, 0, len(drivers))
	for k := range drivers {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

type DriverOptions interface {
	String(key string) string
	Int(key string) int
	Bool(key string) bool
}

// GetDefaultCloudConfig returns a default cloudconfig script
func GetCloudConfig(driver Driver) (string, error) {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".docker/public-key.json"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	key, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	machineOpts, err := driver.GetMachineOptions()
	if err != nil {
		return "", err
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)

	dockerOpts := fmt.Sprintf("DOCKER_OPTS=\"--auth=%s --host unix:// --host=%s --auth-authorized-dir=%s\"",
		machineOpts.Auth, machineOpts.Host, machineOpts.AuthorizedDir)

	buf := bytes.NewBufferString(dockerOpts)

	encodedDockerOpts := base64.StdEncoding.EncodeToString(buf.Bytes())

	cloudConfigOpts := &CloudConfigOptions{
		PublicKey:  encodedKey,
		DockerOpts: encodedDockerOpts,
	}

	// so for the life of me I cannot get the following to work
	// with cloud-init in DigitalOcean ubuntu 14.04.  therefore,
	// I have to do it manually to add the keys.
	//
	var tmpl bytes.Buffer
	t := template.Must(template.New("machine-cloud-config").Parse(`#cloud-config
apt_update: true
apt_sources:
 - source: "deb https://get.docker.com/ubuntu docker main"
   filename: docker.list
   keyserver: keyserver.ubuntu.com
   keyid: A88D21E9

package_update: true

packages:
 - lxc-docker

write_files:
 - encoding: base64
   content: {{.DockerOpts}}
   path: /etc/default/docker
   permissions: 0644

 - encoding: base64
   content: {{.PublicKey}}
   path: /root/.docker/authorized-keys.d/docker-host.json
   permissions: 0600

runcmd:
 - [ curl, -sS, "https://bfirsh.s3.amazonaws.com/docker/docker-1.3.1-dev-identity-auth", "-o", /usr/bin/docker ]
 - [ stop, docker ]
 - [ start, docker ]

final_message: "Docker Machine provisioning complete"
    `))

	if err := t.Execute(&tmpl, cloudConfigOpts); err != nil {
		return "", err
	}

	return tmpl.String(), nil
}
