package driverfactory

import (
	"fmt"

	"github.com/docker/machine/drivers/amazonec2"
	"github.com/docker/machine/drivers/azure"
	"github.com/docker/machine/drivers/digitalocean"
	"github.com/docker/machine/drivers/exoscale"
	"github.com/docker/machine/drivers/generic"
	"github.com/docker/machine/drivers/google"
	"github.com/docker/machine/drivers/hyperv"
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/drivers/openstack"
	"github.com/docker/machine/drivers/rackspace"
	"github.com/docker/machine/drivers/softlayer"
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/drivers/vmwarevcloudair"
	"github.com/docker/machine/drivers/vmwarevsphere"
	"github.com/docker/machine/libmachine/drivers"
)

func NewDriver(driverName, hostName, storePath string) (drivers.Driver, error) {
	var (
		driver drivers.Driver
	)

	switch driverName {
	case "virtualbox":
		driver = virtualbox.NewDriver(hostName, storePath)
	case "digitalocean":
		driver = digitalocean.NewDriver(hostName, storePath)
	case "amazonec2":
		driver = amazonec2.NewDriver(hostName, storePath)
	case "azure":
		driver = azure.NewDriver(hostName, storePath)
	case "exoscale":
		driver = exoscale.NewDriver(hostName, storePath)
	case "generic":
		driver = generic.NewDriver(hostName, storePath)
	case "google":
		driver = google.NewDriver(hostName, storePath)
	case "hyperv":
		driver = hyperv.NewDriver(hostName, storePath)
	case "openstack":
		driver = openstack.NewDriver(hostName, storePath)
	case "rackspace":
		driver = rackspace.NewDriver(hostName, storePath)
	case "softlayer":
		driver = softlayer.NewDriver(hostName, storePath)
	case "vmwarefusion":
		driver = vmwarefusion.NewDriver(hostName, storePath)
	case "vmwarevcloudair":
		driver = vmwarevcloudair.NewDriver(hostName, storePath)
	case "vmwarevsphere":
		driver = vmwarevsphere.NewDriver(hostName, storePath)
	case "none":
		driver = none.NewDriver(hostName, storePath)
	default:
		return nil, fmt.Errorf("Driver %q not recognized", driverName)
	}

	return driver, nil
}
