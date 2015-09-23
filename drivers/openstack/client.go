package openstack

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/version"
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
	Authenticate(d *Driver) error
	InitComputeClient(d *Driver) error
	InitNetworkClient(d *Driver) error

	CreateInstance(d *Driver) (string, error)
	GetInstanceState(d *Driver) (string, error)
	StartInstance(d *Driver) error
	StopInstance(d *Driver) error
	RestartInstance(d *Driver) error
	DeleteInstance(d *Driver) error
	WaitForInstanceStatus(d *Driver, status string) error
	GetInstanceIpAddresses(d *Driver) ([]IpAddress, error)
	CreateKeyPair(d *Driver, name string, publicKey string) error
	DeleteKeyPair(d *Driver, name string) error
	GetNetworkId(d *Driver) (string, error)
	GetFlavorId(d *Driver) (string, error)
	GetImageId(d *Driver) (string, error)
	AssignFloatingIP(d *Driver, floatingIp *FloatingIp, portId string) error
	GetFloatingIPs(d *Driver) ([]FloatingIp, error)
	GetFloatingIpPoolId(d *Driver) (string, error)
	GetInstancePortId(d *Driver) (string, error)
}

type GenericClient struct {
	Provider *gophercloud.ProviderClient
	Compute  *gophercloud.ServiceClient
	Network  *gophercloud.ServiceClient
}

func (c *GenericClient) CreateInstance(d *Driver) (string, error) {
	serverOpts := servers.CreateOpts{
		Name:             d.MachineName,
		FlavorRef:        d.FlavorId,
		ImageRef:         d.ImageId,
		SecurityGroups:   d.SecurityGroups,
		AvailabilityZone: d.AvailabilityZone,
	}
	if d.NetworkId != "" {
		serverOpts.Networks = []servers.Network{
			{
				UUID: d.NetworkId,
			},
		}
	}

	log.Info("Creating machine...")

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

type FloatingIp struct {
	Id        string
	Ip        string
	NetworkId string
	PortId    string
}

func (c *GenericClient) GetInstanceState(d *Driver) (string, error) {
	server, err := c.GetServerDetail(d)
	if err != nil {
		return "", err
	}
	return server.Status, nil
}

func (c *GenericClient) StartInstance(d *Driver) error {
	if result := startstop.Start(c.Compute, d.MachineId); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) StopInstance(d *Driver) error {
	if result := startstop.Stop(c.Compute, d.MachineId); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) RestartInstance(d *Driver) error {
	if result := servers.Reboot(c.Compute, d.MachineId, servers.SoftReboot); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) DeleteInstance(d *Driver) error {
	if result := servers.Delete(c.Compute, d.MachineId); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) WaitForInstanceStatus(d *Driver, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		current, err := servers.Get(c.Compute, d.MachineId).Extract()
		if err != nil {
			return true, err
		}

		if current.Status == "ERROR" {
			return true, fmt.Errorf("Instance creation failed. Instance is in ERROR state")
		}

		if current.Status == status {
			return true, nil
		}

		return false, nil
	}, (d.ActiveTimeout / 4), 4*time.Second)
}

func (c *GenericClient) GetInstanceIpAddresses(d *Driver) ([]IpAddress, error) {
	server, err := c.GetServerDetail(d)
	if err != nil {
		return nil, err
	}
	addresses := []IpAddress{}
	for network, networkAddresses := range server.Addresses {
		for _, element := range networkAddresses.([]interface{}) {
			address := element.(map[string]interface{})

			addr := IpAddress{
				Network: network,
				Address: address["addr"].(string),
			}

			if tp, ok := address["OS-EXT-IPS:type"]; ok {
				addr.AddressType = tp.(string)
			}
			if mac, ok := address["OS-EXT-IPS-MAC:mac_addr"]; ok {
				addr.Mac = mac.(string)
			}

			addresses = append(addresses, addr)
		}
	}
	return addresses, nil
}

func (c *GenericClient) GetNetworkId(d *Driver) (string, error) {
	return c.getNetworkId(d, d.NetworkName)
}

func (c *GenericClient) GetFloatingIpPoolId(d *Driver) (string, error) {
	return c.getNetworkId(d, d.FloatingIpPool)
}

func (c *GenericClient) getNetworkId(d *Driver, networkName string) (string, error) {
	opts := networks.ListOpts{Name: networkName}
	pager := networks.List(c.Network, opts)
	networkId := ""

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, n := range networkList {
			if n.Name == networkName {
				networkId = n.ID
				return false, nil
			}
		}

		return true, nil
	})

	return networkId, err
}

func (c *GenericClient) GetFlavorId(d *Driver) (string, error) {
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

func (c *GenericClient) GetImageId(d *Driver) (string, error) {
	opts := images.ListOpts{Name: d.ImageName}
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
	if result := keypairs.Delete(c.Compute, name); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) GetServerDetail(d *Driver) (*servers.Server, error) {
	server, err := servers.Get(c.Compute, d.MachineId).Extract()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (c *GenericClient) AssignFloatingIP(d *Driver, floatingIp *FloatingIp, portId string) error {
	if floatingIp.Id == "" {
		f, err := floatingips.Create(c.Network, floatingips.CreateOpts{
			FloatingNetworkID: d.FloatingIpPoolId,
			PortID:            portId,
		}).Extract()
		if err != nil {
			return err
		}
		floatingIp.Id = f.ID
		floatingIp.Ip = f.FloatingIP
		floatingIp.NetworkId = f.FloatingNetworkID
		floatingIp.PortId = f.PortID
		return nil
	}
	_, err := floatingips.Update(c.Network, floatingIp.Id, floatingips.UpdateOpts{
		PortID: portId,
	}).Extract()
	if err != nil {
		return err
	}
	return nil
}

func (c *GenericClient) GetFloatingIPs(d *Driver) ([]FloatingIp, error) {
	pager := floatingips.List(c.Network, floatingips.ListOpts{
		FloatingNetworkID: d.FloatingIpPoolId,
	})

	ips := []FloatingIp{}
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		floatingipList, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}
		for _, f := range floatingipList {
			ips = append(ips, FloatingIp{
				Id:        f.ID,
				Ip:        f.FloatingIP,
				NetworkId: f.FloatingNetworkID,
				PortId:    f.PortID,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}
	return ips, nil
}

func (c *GenericClient) GetInstancePortId(d *Driver) (string, error) {
	pager := ports.List(c.Network, ports.ListOpts{
		DeviceID:  d.MachineId,
		NetworkID: d.NetworkId,
	})

	var portId string
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		portList, err := ports.ExtractPorts(page)
		if err != nil {
			return false, err
		}
		for _, port := range portList {
			portId = port.ID
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return "", err
	}
	return portId, nil
}

func (c *GenericClient) InitComputeClient(d *Driver) error {
	if c.Compute != nil {
		return nil
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

func (c *GenericClient) InitNetworkClient(d *Driver) error {
	if c.Network != nil {
		return nil
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
	if c.Provider != nil {
		return nil
	}

	log.WithFields(log.Fields{
		"AuthUrl":    d.AuthUrl,
		"Insecure":   d.Insecure,
		"DomainID":   d.DomainID,
		"DomainName": d.DomainName,
		"Username":   d.Username,
		"TenantName": d.TenantName,
		"TenantID":   d.TenantId,
	}).Debug("Authenticating...")

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: d.AuthUrl,
		DomainID:         d.DomainID,
		DomainName:       d.DomainName,
		Username:         d.Username,
		Password:         d.Password,
		TenantName:       d.TenantName,
		TenantID:         d.TenantId,
		AllowReauth:      true,
	}

	provider, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return err
	}

	provider.UserAgent.Prepend(fmt.Sprintf("docker-machine/v%d", version.ApiVersion))

	if d.Insecure {
		// Configure custom TLS settings.
		config := &tls.Config{InsecureSkipVerify: true}
		transport := &http.Transport{TLSClientConfig: config}
		provider.HTTPClient.Transport = transport
	}

	err = openstack.Authenticate(provider, opts)
	if err != nil {
		return err
	}

	c.Provider = provider

	return nil
}
