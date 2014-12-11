package rackspace

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers/openstack"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace"
)

func unsupportedOpErr(operation string) error {
	return fmt.Errorf("Rackspace does not currently support the %s operation", operation)
}

// Client is a Rackspace specialization of the generic OpenStack driver.
type Client struct {
	openstack.GenericClient

	driver *Driver
}

// Authenticate creates a Rackspace-specific Gophercloud client.
func (c *Client) Authenticate(d *openstack.Driver) error {
	log.WithFields(log.Fields{
		"Username": d.Username,
	}).Info("Authenticating...")

	apiKey := c.driver.APIKey
	opts := gophercloud.AuthOptions{
		Username: d.Username,
		APIKey:   apiKey,
	}

	provider, err := rackspace.AuthenticatedClient(opts)
	if err != nil {
		return err
	}
	c.Provider = provider

	return nil
}

// StartInstance is unfortunately not supported on Rackspace at this time.
func (c *Client) StartInstance(d *openstack.Driver) error {
	return unsupportedOpErr("start")
}

// StopInstance is unfortunately not support on Rackspace at this time.
func (c *Client) StopInstance(d *openstack.Driver) error {
	return unsupportedOpErr("stop")
}
