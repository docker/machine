package rackspace

import (
	"os"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/openstack"
)

// Driver is a machine driver for Rackspace. It's a specialization of the generic OpenStack one.
type Driver struct {
	drivers.Driver

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
	createFlags.EndpointType = cmd.String(
		[]string{"-rackspace-endpoint-type"},
		os.Getenv("OS_ENDPOINT_TYPE"),
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
