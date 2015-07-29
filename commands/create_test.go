package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSwarmDiscoveryErrorsGivenInvalidURL(t *testing.T) {
	err := validateSwarmDiscovery("foo")
	assert.Error(t, err)
}

func TestValidateSwarmDiscoveryAcceptsEmptyString(t *testing.T) {
	err := validateSwarmDiscovery("")
	assert.NoError(t, err)
}

func TestValidateSwarmDiscoveryAcceptsValidFormat(t *testing.T) {
	err := validateSwarmDiscovery("token://deadbeefcafe")
	assert.NoError(t, err)
}
