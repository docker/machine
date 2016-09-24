package azure

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestParseSecurityRuleProtocol(t *testing.T) {
	tests := []struct {
		raw           string
		expectedProto network.SecurityRuleProtocol
		expectedErr   bool
	}{
		{"tcp", network.TCP, false},
		{"udp", network.UDP, false},
		{"*", network.Asterisk, false},
		{"Invalid", "", true},
	}

	for _, tc := range tests {
		proto, err := parseSecurityRuleProtocol(tc.raw)
		assert.Equal(t, tc.expectedProto, proto)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestGetSecurityRules(t *testing.T) {
	d := Driver{BaseDriver: &drivers.BaseDriver{SSHPort: 22}, DockerPort: 2376}
	expected := []struct {
		name     string
		protocol network.SecurityRuleProtocol
		port     string
	}{
		{"SSHAllowAny", network.TCP, "22"},
		{"DockerAllowAny", network.TCP, "2376"},
		{"Port123TcpAllowAny", network.TCP, "123"},
		{"Port123UdpAllowAny", network.UDP, "123"},
	}
	rules, err := d.getSecurityRules([]string{"123/tcp", "123/udp"})
	assert.Nil(t, err)
	for i, r := range *rules {
		assert.EqualValues(t, expected[i].name, *r.Name)
		assert.EqualValues(t, expected[i].protocol, r.Properties.Protocol)
		assert.EqualValues(t, expected[i].port, *r.Properties.DestinationPortRange)
	}
}
