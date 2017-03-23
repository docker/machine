package generic

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"generic-ip-address": "localhost",
			"generic-ssh-key":    "path",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestGenericIPAddressError(t *testing.T) {
	driver := NewDriver("default2", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"generic-ssh-key": "path",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.EqualError(
		t,
		err,
		"generic driver requires the --generic-ip-address option",
		"SetConfigFromFlags should throw an error when generic-ip-address is missing",
	)
}

func TestGenericDockerPortDeprecationError(t *testing.T) {
	driver := NewDriver("default3", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"generic-ip-address":  "localhost",
			"generic-ssh-key":     "sshkey",
			"generic-engine-port": 12345,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.EqualError(
		t,
		err,
		"-generic-engine-port has been deprecated in favor of: -engine-port",
		"SetConfigFromFlags should throw an error when generic-engine-port is set",
	)
}
