package rackspace

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/openstack"
)

// Driver is a machine driver for Rackspace. It's a specialization of the generic OpenStack one.
type Driver struct {
	*openstack.Driver

	APIKey string
}

// CreateFlags stores the command-line arguments given to "machine create".
type CreateFlags struct {
	Username     *string
	APIKey       *string
	Region       *string
	MachineName  *string
	EndpointType *string
	ImageID      *string
	FlavorID     *string
	SSHUser      *string
	SSHPort      *int
}

func init() {
	drivers.Register("rackspace", &drivers.RegisteredDriver{
		New:                 NewDriver,
		RegisterCreateFlags: RegisterCreateFlags,
	})
}

// RegisterCreateFlags registers the "machine create" flags recognized by this driver, including
// their help text and defaults.
func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
	createFlags := new(CreateFlags)
	createFlags.Username = cmd.String(
		[]string{"-rackspace-username"},
		os.Getenv("OS_USERNAME"),
		"Rackspace account username",
	)
	createFlags.APIKey = cmd.String(
		[]string{"-rackspace-api-key"},
		os.Getenv("OS_API_KEY"),
		"Rackspace API key",
	)
	createFlags.Region = cmd.String(
		[]string{"-rackspace-region"},
		os.Getenv("OS_REGION_NAME"),
		"Rackspace region name",
	)
	endpointDefault := os.Getenv("OS_ENDPOINT_TYPE")
	if endpointDefault == "" {
		endpointDefault = "publicURL"
	}
	createFlags.EndpointType = cmd.String(
		[]string{"-rackspace-endpoint-type"},
		endpointDefault,
		"Rackspace endpoint type (adminURL, internalURL or the default publicURL)",
	)
	createFlags.ImageID = cmd.String(
		[]string{"-rackspace-image-id"},
		"",
		"Rackspace image ID. Default: Ubuntu 14.10 (Utopic Unicorn) (PVHVM)",
	)
	createFlags.FlavorID = cmd.String(
		[]string{"-rackspace-flavor-id"},
		"general1-1",
		"Rackspace flavor ID. Default: General Purpose 1GB",
	)
	createFlags.SSHUser = cmd.String(
		[]string{"-rackspace-ssh-user"},
		"root",
		"SSH user for the newly booted machine. Set to root by default",
	)
	createFlags.SSHPort = cmd.Int(
		[]string{"-rackspace-ssh-port"},
		22,
		"SSH port for the newly booted machine. Set to 22 by default",
	)
	return createFlags
}

// NewDriver instantiates a Rackspace driver.
func NewDriver(storePath string) (drivers.Driver, error) {
	log.WithFields(log.Fields{
		"storePath": storePath,
	}).Info("Instantiating Rackspace driver.")

	client := &Client{}
	inner, err := openstack.NewDerivedDriver(storePath, &Client{})
	if err != nil {
		return nil, err
	}

	driver := &Driver{Driver: inner}
	client.driver = driver
	return driver, nil
}

// DriverName is the user-visible name of this driver.
func (d *Driver) DriverName() string {
	return "rackspace"
}

// SetConfigFromFlags assigns and verifies the command-line arguments presented to the driver.
func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
	flags := flagsInterface.(*CreateFlags)

	d.Username = *flags.Username
	d.APIKey = *flags.APIKey
	d.Region = *flags.Region
	d.MachineName = *flags.MachineName
	d.EndpointType = *flags.EndpointType
	d.ImageId = *flags.ImageID
	d.FlavorId = *flags.FlavorID
	d.SSHUser = *flags.SSHUser
	d.SSHPort = *flags.SSHPort

	return d.checkConfig()
}

func missingEnvOrOption(setting, envVar, opt string) error {
	return fmt.Errorf(
		"%s must be specified either using the environment variable %s or the CLI option %s",
		setting,
		envVar,
		opt,
	)
}

func (d *Driver) checkConfig() error {
	if d.Username == "" {
		return missingEnvOrOption("Username", "OS_USERNAME", "--rackspace-username")
	}
	if d.APIKey == "" {
		return missingEnvOrOption("API key", "OS_API_KEY", "--rackspace-api-key")
	}

	if d.ImageId == "" {
		// Default to the Ubuntu 14.10 image.
		// This is done here, rather than in the option registration, to keep the default value
		// from making "machine create --help" ugly.
		d.ImageId = "0766e5df-d60a-4100-ae8c-07f27ec0148f"
	}

	if d.EndpointType != "publicURL" && d.EndpointType != "adminURL" && d.EndpointType != "internalURL" {
		return fmt.Errorf(`Invalid endpoint type "%s". Endpoint type must be publicURL, adminURL or internalURL.`, d.EndpointType)
	}

	return nil
}
