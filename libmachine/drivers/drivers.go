package drivers

import (
	"errors"
	"fmt"
	"sort"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/state"
)

// Driver defines how a host is created and controlled. Different types of
// driver represent different ways hosts can be created (e.g. different
// hypervisors, different cloud providers)
type Driver interface {
	// Create a host using the driver's config
	Create() error

	// DriverName returns the name of the driver as it is registered
	DriverName() string

	// GetIP returns an IP or hostname that this host is available at
	// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
	GetIP() (string, error)

	// GetMachineName returns the name of the machine
	GetMachineName() string

	// GetSSHHostname returns hostname for use with ssh
	GetSSHHostname() (string, error)

	// GetSSHKeyPath returns key path for use with ssh
	GetSSHKeyPath() string

	// GetSSHPort returns port for use with ssh
	GetSSHPort() (int, error)

	// GetSSHUsername returns username for use with ssh
	GetSSHUsername() string

	// GetURL returns a Docker compatible host URL for connecting to this host
	// e.g. tcp://1.2.3.4:2376
	GetURL() (string, error)

	// GetState returns the state that the host is in (running, stopped, etc)
	GetState() (state.State, error)

	// Kill stops a host forcefully
	Kill() error

	// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
	PreCreateCheck() error

	// Remove a host
	Remove() error

	// Restart a host. This may just call Stop(); Start() if the provider does not
	// have any special restart behaviour.
	Restart() error

	// SetConfigFromFlags configures the driver with the object that was returned
	// by RegisterCreateFlags
	SetConfigFromFlags(flags DriverOptions) error

	// Start a host
	Start() error

	// Stop a host gracefully
	Stop() error
}

// RegisteredDriver is a wrapper struct used to "register" drivers so that
// information about them can be displayed in the CLI.
type RegisteredDriver struct {
	// GetCreateFlags gets the FlagSet for "docker-machine create" for
	// display to the user.
	GetCreateFlags func() []cli.Flag
}

var ErrHostIsNotRunning = errors.New("Host is not running")

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

func GetCreateFlagsForDriver(name string) ([]cli.Flag, error) {

	for driverName := range drivers {
		if name == driverName {
			driver := drivers[driverName]
			flags := driver.GetCreateFlags()
			sort.Sort(ByFlagName(flags))
			return flags, nil
		}
	}

	return nil, fmt.Errorf("Driver %s not found", name)
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
	StringSlice(key string) []string
	Int(key string) int
	Bool(key string) bool
}

func MachineInState(d Driver, desiredState state.State) func() bool {
	return func() bool {
		currentState, err := d.GetState()
		if err != nil {
			log.Debugf("Error getting machine state: %s", err)
		}
		if currentState == desiredState {
			return true
		}
		return false
	}
}
