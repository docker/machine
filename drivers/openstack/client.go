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
	GetInstanceIPAddresses(d *Driver) ([]IPAddress, error)
	CreateKeyPair(d *Driver, name string, publicKey string) error
	DeleteKeyPair(d *Driver, name string) error
	GetNetworkID(d *Driver) (string, error)
	GetFlavorID(d *Driver) (string, error)
	GetImageID(d *Driver) (string, error)
	AssignFloatingIP(d *Driver, floatingIP *FloatingIP, portID string) error
	GetFloatingIPs(d *Driver) ([]FloatingIP, error)
	GetFloatingIPPoolID(d *Driver) (string, error)
	GetInstancePortID(d *Driver) (string, error)
}

type GenericClient struct {
	Provider *gophercloud.ProviderClient
	Compute  *gophercloud.ServiceClient
	Network  *gophercloud.ServiceClient
}

func (c *GenericClient) CreateInstance(d *Driver) (string, error) {
	serverOpts := servers.CreateOpts{
		Name:             d.MachineName,
		FlavorRef:        d.FlavorID,
		ImageRef:         d.ImageID,
		SecurityGroups:   d.SecurityGroups,
		AvailabilityZone: d.AvailabilityZone,
	}
	if d.NetworkID != "" {
		serverOpts.Networks = []servers.Network{
			{
				UUID: d.NetworkID,
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

type IPAddress struct {
	Network     string
	AddressType string
	Address     string
	Version     int
	Mac         string
}

type FloatingIP struct {
	ID        string
	IP        string
	NetworkID string
	PortID    string
}

func (c *GenericClient) GetInstanceState(d *Driver) (string, error) {
	server, err := c.GetServerDetail(d)
	if err != nil {
		return "", err
	}
	return server.Status, nil
}

func (c *GenericClient) StartInstance(d *Driver) error {
	if result := startstop.Start(c.Compute, d.MachineID); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) StopInstance(d *Driver) error {
	if result := startstop.Stop(c.Compute, d.MachineID); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) RestartInstance(d *Driver) error {
	if result := servers.Reboot(c.Compute, d.MachineID, servers.SoftReboot); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) DeleteInstance(d *Driver) error {
	if result := servers.Delete(c.Compute, d.MachineID); result.Err != nil {
		return result.Err
	}
	return nil
}

func (c *GenericClient) WaitForInstanceStatus(d *Driver, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		current, err := servers.Get(c.Compute, d.MachineID).Extract()
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

func (c *GenericClient) GetInstanceIPAddresses(d *Driver) ([]IPAddress, error) {
	server, err := c.GetServerDetail(d)
	if err != nil {
		return nil, err
	}
	addresses := []IPAddress{}
	for network, networkAddresses := range server.Addresses {
		for _, element := range networkAddresses.([]interface{}) {
			address := element.(map[string]interface{})
			version, ok := address["version"].(float64)
			if !ok {
				// Assume IPv4 if no version present.
				version = 4
			}

			addr := IPAddress{
				Network:     network,
				AddressType: Fixed,
				Address:     address["addr"].(string),
				Version:     int(version),
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

func (c *GenericClient) GetNetworkID(d *Driver) (string, error) {
	return c.getNetworkID(d, d.NetworkName)
}

func (c *GenericClient) GetFloatingIPPoolID(d *Driver) (string, error) {
	return c.getNetworkID(d, d.FloatingIPPool)
}

func (c *GenericClient) getNetworkID(d *Driver, networkName string) (string, error) {
	opts := networks.ListOpts{Name: networkName}
	pager := networks.List(c.Network, opts)
	networkID := ""

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		networkList, err := networks.ExtractNetworks(page)
		if err != nil {
			return false, err
		}

		for _, n := range networkList {
			if n.Name == networkName {
				networkID = n.ID
				return false, nil
			}
		}

		return true, nil
	})

	return networkID, err
}

func (c *GenericClient) GetFlavorID(d *Driver) (string, error) {
	pager := flavors.ListDetail(c.Compute, nil)
	flavorID := ""

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}

		for _, f := range flavorList {
			if f.Name == d.FlavorName {
				flavorID = f.ID
				return false, nil
			}
		}

		return true, nil
	})

	return flavorID, err
}

func (c *GenericClient) GetImageID(d *Driver) (string, error) {
	opts := images.ListOpts{Name: d.ImageName}
	pager := images.ListDetail(c.Compute, opts)
	imageID := ""

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, i := range imageList {
			if i.Name == d.ImageName {
				imageID = i.ID
				return false, nil
			}
		}

		return true, nil
	})

	return imageID, err
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
	server, err := servers.Get(c.Compute, d.MachineID).Extract()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (c *GenericClient) AssignFloatingIP(d *Driver, floatingIP *FloatingIP, portID string) error {
	if floatingIP.ID == "" {
		f, err := floatingips.Create(c.Network, floatingips.CreateOpts{
			FloatingNetworkID: d.FloatingIPPoolID,
			PortID:            portID,
		}).Extract()
		if err != nil {
			return err
		}
		floatingIP.ID = f.ID
		floatingIP.IP = f.FloatingIP
		floatingIP.NetworkID = f.FloatingNetworkID
		floatingIP.PortID = f.PortID
		return nil
	}
	_, err := floatingips.Update(c.Network, floatingIP.ID, floatingips.UpdateOpts{
		PortID: portID,
	}).Extract()
	if err != nil {
		return err
	}
	return nil
}

func (c *GenericClient) GetFloatingIPs(d *Driver) ([]FloatingIP, error) {
	pager := floatingips.List(c.Network, floatingips.ListOpts{
		FloatingNetworkID: d.FloatingIPPoolID,
	})

	ips := []FloatingIP{}
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		floatingipList, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err
		}
		for _, f := range floatingipList {
			ips = append(ips, FloatingIP{
				ID:        f.ID,
				IP:        f.FloatingIP,
				NetworkID: f.FloatingNetworkID,
				PortID:    f.PortID,
			})
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}
	return ips, nil
}

func (c *GenericClient) GetInstancePortID(d *Driver) (string, error) {
	pager := ports.List(c.Network, ports.ListOpts{
		DeviceID:  d.MachineID,
		NetworkID: d.NetworkID,
	})

	var portID string
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		portList, err := ports.ExtractPorts(page)
		if err != nil {
			return false, err
		}
		for _, port := range portList {
			portID = port.ID
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return "", err
	}
	return portID, nil
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
		"AuthUrl":    d.AuthURL,
		"Insecure":   d.Insecure,
		"DomainID":   d.DomainID,
		"DomainName": d.DomainName,
		"Username":   d.Username,
		"TenantName": d.TenantName,
		"TenantID":   d.TenantID,
	}).Debug("Authenticating...")

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: d.AuthURL,
		DomainID:         d.DomainID,
		DomainName:       d.DomainName,
		Username:         d.Username,
		Password:         d.Password,
		TenantName:       d.TenantName,
		TenantID:         d.TenantID,
		AllowReauth:      true,
	}

	provider, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return err
	}

	provider.UserAgent.Prepend(fmt.Sprintf("docker-machine/v%d", version.APIVersion))

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
