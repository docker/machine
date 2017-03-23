package azure

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestGenericDockerPortDeprecationError(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"azure-subscription-id": "abcdef",
			"azure-docker-port":     12345,
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.EqualError(
		t,
		err,
		"-azure-docker-port has been deprecated in favor of: -engine-port",
		"SetConfigFromFlags should throw an error when generic-docker-port is set",
	)
}
