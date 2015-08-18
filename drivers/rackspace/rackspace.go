package rackspace

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers/openstack"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
)

// Driver is a machine driver for Rackspace. It's a specialization of the generic OpenStack one.
type Driver struct {
	*openstack.Driver

	APIKey string
}

const (
	defaultEndpointType  = "publicURL"
	defaultFlavorId      = "general1-1"
	defaultSSHUser       = "root"
	defaultSSHPort       = 22
	defaultDockerInstall = "true"
)

func init() {
	drivers.Register("rackspace", &drivers.RegisteredDriver{
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the "machine create" flags recognized by this driver, including
// their help text and defaults.
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "OS_USERNAME",
			Name:   "rackspace-username",
			Usage:  "Rackspace account username",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_API_KEY",
			Name:   "rackspace-api-key",
			Usage:  "Rackspace API key",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_REGION_NAME",
			Name:   "rackspace-region",
			Usage:  "Rackspace region name",
			Value:  "",
		},
		cli.StringFlag{
			EnvVar: "OS_ENDPOINT_TYPE",
			Name:   "rackspace-endpoint-type",
			Usage:  "Rackspace endpoint type (adminURL, internalURL or the default publicURL)",
			Value:  defaultEndpointType,
		},
		cli.StringFlag{
			Name:  "rackspace-image-id",
			Usage: "Rackspace image ID. Default: Ubuntu 14.04 LTS (Trusty Tahr) (PVHVM)",
		},
		cli.StringFlag{
			Name:   "rackspace-flavor-id",
			Usage:  "Rackspace flavor ID. Default: General Purpose 1GB",
			Value:  defaultFlavorId,
			EnvVar: "OS_FLAVOR_ID",
		},
		cli.StringFlag{
			Name:  "rackspace-ssh-user",
			Usage: "SSH user for the newly booted machine. Set to root by default",
			Value: defaultSSHUser,
		},
		cli.IntFlag{
			Name:  "rackspace-ssh-port",
			Usage: "SSH port for the newly booted machine. Set to 22 by default",
			Value: defaultSSHPort,
		},
		cli.StringFlag{
			Name:  "rackspace-docker-install",
			Usage: "Set if docker have to be installed on the machine",
			Value: defaultDockerInstall,
		},
	}
}

// NewDriver instantiates a Rackspace driver.
func NewDriver(machineName, storePath string) drivers.Driver {
	log.WithFields(log.Fields{
		"machineName": machineName,
	}).Debug("Instantiating Rackspace driver.")

	inner := openstack.NewDerivedDriver(machineName, storePath)

	return &Driver{
		Driver: inner,
	}
}

// DriverName is the user-visible name of this driver.
func (d *Driver) DriverName() string {
	return "rackspace"
}

func missingEnvOrOption(setting, envVar, opt string) error {
	return fmt.Errorf(
		"%s must be specified either using the environment variable %s or the CLI option %s",
		setting,
		envVar,
		opt,
	)
}

// SetConfigFromFlags assigns and verifies the command-line arguments presented to the driver.
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Username = flags.String("rackspace-username")
	d.APIKey = flags.String("rackspace-api-key")
	d.Region = flags.String("rackspace-region")
	d.EndpointType = flags.String("rackspace-endpoint-type")
	d.ImageId = flags.String("rackspace-image-id")
	d.FlavorId = flags.String("rackspace-flavor-id")
	d.SSHUser = flags.String("rackspace-ssh-user")
	d.SSHPort = flags.Int("rackspace-ssh-port")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	if d.Region == "" {
		return missingEnvOrOption("Region", "OS_REGION_NAME", "--rackspace-region")
	}
	if d.Username == "" {
		return missingEnvOrOption("Username", "OS_USERNAME", "--rackspace-username")
	}
	if d.APIKey == "" {
		return missingEnvOrOption("API key", "OS_API_KEY", "--rackspace-api-key")
	}

	if d.ImageId == "" {
		// Default to the Ubuntu 14.04 image.
		// This is done here, rather than in the option registration, to keep the default value
		// from making "machine create --help" ugly.
		d.ImageId = "598a4282-f14b-4e50-af4c-b3e52749d9f9"
	}

	if d.EndpointType != "publicURL" && d.EndpointType != "adminURL" && d.EndpointType != "internalURL" {
		return fmt.Errorf(`Invalid endpoint type "%s". Endpoint type must be publicURL, adminURL or internalURL.`, d.EndpointType)
	}

	return nil
}
