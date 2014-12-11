package openstack

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/rackspace/gophercloud/openstack/compute/v2/flavors"
	"github.com/rackspace/gophercloud/openstack/compute/v2/images"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/rackspace/gophercloud/openstack/networking/v2/networks"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"github.com/rackspace/gophercloud/pagination"
)

type Client interface {
	CreateInstance(d *Driver) (string, error)
	GetInstanceState(d *Driver) (string, error)
	StartInstance(d *Driver) error
	StopInstance(d *Driver) error
	RestartInstance(d *Driver) error
	DeleteInstance(d *Driver) error
	WaitForInstanceStatus(d *Driver, status string, timeout int) error
	GetInstanceIpAddresses(d *Driver) ([]IpAddress, error)
	CreateKeyPair(d *Driver, name string, publicKey string) error
	DeleteKeyPair(d *Driver, name string) error
	Authenticate(d *Driver) error
}

type GenericClient struct {
	Provider *gophercloud.ProviderClient
	Compute  *gophercloud.ServiceClient
	Network  *gophercloud.ServiceClient
}

func (c *GenericClient) CreateInstance(d *Driver) (string, error) {
	if err := c.initComputeClient(d); err != nil {
		return "", err
	}
	serverOpts := servers.CreateOpts{
		Name:           d.MachineName,
		FlavorRef:      d.FlavorId,
		ImageRef:       d.ImageId,
		SecurityGroups: d.SecurityGroups,
	}
	if d.NetworkId != "" {
		serverOpts.Networks = []servers.Network{
			{
				UUID: d.NetworkId,
			},
		}
	}
	server, err := servers.Create(c.Compute, keypairs.CreateOptsExt{
		serverOpts,
		d.KeyPairName,
	}).Extract()
	if err != nil {
		return "", err
	}
	return server.ID, nil
}

const (
	Floating string = "floating"
	Fixed    string = "fixed"
)

type IpAddress struct {
	Network     string
	AddressType string
	Address     string
	Mac         string
}

func (c *GenericClient) GetInstanceState(d *Driver) (string, error) {
	server, err := c.getServerDetail(d)
	if err != nil {
		return "", err
	}

	c.getFloatingIPs(d)

	c.getPorts(d)

	return server.Status, nil
}

func (c *GenericClient) StartInstance(d *Driver) error {
	if err := c.initComputeClient(d); err != nil {
		return err
	}
	if result := startstop.Start(c.Compute, d.MachineId); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) StopInstance(d *Driver) error {
	if err := c.initComputeClient(d); err != nil {
		return err
	}
	if result := startstop.Stop(c.Compute, d.MachineId); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) RestartInstance(d *Driver) error {
	if err := c.initComputeClient(d); err != nil {
		return err
	}
	if result := servers.Reboot(c.Compute, d.MachineId, servers.SoftReboot); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) DeleteInstance(d *Driver) error {
	if err := c.initComputeClient(d); err != nil {
		return err
	}
	if result := servers.Delete(c.Compute, d.MachineId); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) WaitForInstanceStatus(d *Driver, status string, timeout int) error {
	if err := servers.WaitForStatus(c.Compute, d.MachineId, status, timeout); err != nil {
		return err
	}
	return nil
}

func (c *GenericClient) GetInstanceIpAddresses(d *Driver) ([]IpAddress, error) {
	server, err := c.getServerDetail(d)
	if err != nil {
		return nil, err
	}
	addresses := []IpAddress{}
	for network, networkAddresses := range server.Addresses {
		for _, element := range networkAddresses.([]interface{}) {
			address := element.(map[string]interface{})
			addresses = append(addresses, IpAddress{
				Network:     network,
				AddressType: address["OS-EXT-IPS:type"].(string),
				Address:     address["addr"].(string),
				Mac:         address["OS-EXT-IPS-MAC:mac_addr"].(string),
			})
		}
	}
	return addresses, nil
}

func (c *GenericClient) GetNetworkId(d *Driver, networkName string) (string, error) {
	if err := c.initNetworkClient(d); err != nil {
		return "", err
	}

	opts := networks.ListOpts{Name: d.NetworkName}
	pager := networks.List(c.Network, opts)
	networkId := ""

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, n := range networkList {
			if n.Name == d.NetworkName {
				networkId = n.ID
				return false, nil
			}
		}

		return true, nil
	})

	return networkId, err
}

func (c *GenericClient) GetFlavorId(d *Driver, flavorName string) (string, error) {
	if err := c.initComputeClient(d); err != nil {
		return "", err
	}

	pager := flavors.ListDetail(c.Compute, nil)
	flavorId := ""

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}

		for _, f := range flavorList {
			if f.Name == d.FlavorName {
				flavorId = f.ID
				return false, nil
			}
		}

		return true, nil
	})

	return flavorId, err
}

func (c *GenericClient) GetImageId(d *Driver, imageName string) (string, error) {
	if err := c.initComputeClient(d); err != nil {
		return "", err
	}

	opts := images.ListOpts{Name: imageName}
	pager := images.ListDetail(c.Compute, opts)
	imageId := ""

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, i := range imageList {
			if i.Name == d.ImageName {
				imageId = i.ID
				return false, nil
			}
		}

		return true, nil
	})

	return imageId, err
}

func (c *GenericClient) CreateKeyPair(d *Driver, name string, publicKey string) error {
	if err := c.initComputeClient(d); err != nil {
		return err
	}
	opts := keypairs.CreateOpts{
		Name:      name,
		PublicKey: publicKey,
	}
	if result := keypairs.Create(c.Compute, opts); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) DeleteKeyPair(d *Driver, name string) error {
	if err := c.initComputeClient(d); err != nil {
		return err
	}
	if result := keypairs.Delete(c.Compute, name); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) getServerDetail(d *Driver) (*servers.Server, error) {
	if err := c.initComputeClient(d); err != nil {
		return nil, err
	}
	server, err := servers.Get(c.Compute, d.MachineId).Extract()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (c *GenericClient) getFloatingIPs(d *Driver) ([]string, error) {

	if err := c.initNetworkClient(d); err != nil {
		return nil, err
	}

	pager := floatingips.List(c.Network, floatingips.ListOpts{})

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		floatingipList, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}
		for _, f := range floatingipList {
			log.Debug("### FloatingIP => %s", f)
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *GenericClient) getPorts(d *Driver) ([]string, error) {

	if err := c.initNetworkClient(d); err != nil {
		return nil, err
	}

	pager := ports.List(c.Network, ports.ListOpts{
		DeviceID: d.MachineId,
	})

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		portList, err := ports.ExtractPorts(page)
		if err != nil {
			return false, err
		}
		for _, port := range portList {
			log.Debug("### Port => %s", port)
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *GenericClient) initComputeClient(d *Driver) error {
	if c.Provider == nil {
		err := c.Authenticate(d)
		if err != nil {
			return err
		}
	}
	compute, err := openstack.NewComputeV2(c.Provider, gophercloud.EndpointOpts{
		Region:       d.Region,
		Availability: c.getEndpointType(d),
	})
	if err != nil {
		return err
	}
	c.Compute = compute
	return nil
}

func (c *GenericClient) initNetworkClient(d *Driver) error {
	if c.Provider == nil {
		err := c.Authenticate(d)
		if err != nil {
			return err
		}
	}
	network, err := openstack.NewNetworkV2(c.Provider, gophercloud.EndpointOpts{
		Region:       d.Region,
		Availability: c.getEndpointType(d),
	})
	if err != nil {
		return err
	}
	c.Network = network
	return nil
}

func (c *GenericClient) getEndpointType(d *Driver) gophercloud.Availability {
	switch d.EndpointType {
	case "internalURL":
		return gophercloud.AvailabilityInternal
	case "adminURL":
		return gophercloud.AvailabilityAdmin
	}
	return gophercloud.AvailabilityPublic
}

func (c *GenericClient) Authenticate(d *Driver) error {
	log.WithFields(log.Fields{
		"AuthUrl":    d.AuthUrl,
		"Username":   d.Username,
		"TenantName": d.TenantName,
		"TenantID":   d.TenantId,
	}).Debug("Authenticating...")

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: d.AuthUrl,
		Username:         d.Username,
		Password:         d.Password,
		TenantName:       d.TenantName,
		TenantID:         d.TenantId,
		AllowReauth:      true,
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return err
	}
	c.Provider = provider

	return nil
}
