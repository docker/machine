package drivers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP(t *testing.T) {
	cases := []struct {
		baseDriver  *BaseDriver
		expectedIP  string
		expectedErr error
	}{
		{&BaseDriver{}, "", errors.New("IP address is not set")},
		{&BaseDriver{IPAddress: "2001:4860:0:2001::68"}, "2001:4860:0:2001::68", nil},
		{&BaseDriver{IPAddress: "192.168.0.1"}, "192.168.0.1", nil},
		{&BaseDriver{IPAddress: "::1"}, "::1", nil},
		{&BaseDriver{IPAddress: "hostname"}, "hostname", nil},
	}

	for _, c := range cases {
		ip, err := c.baseDriver.GetIP()
		assert.Equal(t, c.expectedIP, ip)
		assert.Equal(t, c.expectedErr, err)
	}
}
