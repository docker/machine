package openstack

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"openstack-auth-url":  "http://url",
			"openstack-username":  "user",
			"openstack-password":  "pwd",
			"openstack-tenant-id": "ID",
			"openstack-flavor-id": "ID",
			"openstack-image-id":  "ID",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetSingleNetworkId(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"openstack-auth-url":  "http://url",
			"openstack-username":  "user",
			"openstack-password":  "pwd",
			"openstack-tenant-id": "ID",
			"openstack-flavor-id": "ID",
			"openstack-image-id":  "ID",
			"openstack-net-id":    "ID",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetSingleNetworkName(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"openstack-auth-url":  "http://url",
			"openstack-username":  "user",
			"openstack-password":  "pwd",
			"openstack-tenant-id": "ID",
			"openstack-flavor-id": "ID",
			"openstack-image-id":  "ID",
			"openstack-net-name":  "ID",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetMultipleNetworkIds(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"openstack-auth-url":  "http://url",
			"openstack-username":  "user",
			"openstack-password":  "pwd",
			"openstack-tenant-id": "ID",
			"openstack-flavor-id": "ID",
			"openstack-image-id":  "ID",
			//TODO: multivalue test
			//"openstack-net-id":    "ID",
			"openstack-net-id": "ID2",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetMultipleNetworkNames(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"openstack-auth-url":  "http://url",
			"openstack-username":  "user",
			"openstack-password":  "pwd",
			"openstack-tenant-id": "ID",
			"openstack-flavor-id": "ID",
			"openstack-image-id":  "ID",
			"openstack-net-name":  "ID",
			//TODO: multivalue test
			//"openstack-net-name":  "ID2",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}
