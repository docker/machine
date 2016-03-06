package azure

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestCreate(t *testing.T) {
	driver := NewDriver("azure-vm-cxtest", "<enter path to your docker-machine directory>")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"azure-subscription-cert": "<enter your subscript cert path>",
			"azure-subscription-id":   "<subid>",
		},
		CreateFlags: driver.GetCreateFlags(),
	}
	err := driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)

	err = driver.Create()
	assert.NoError(t, err)
}
