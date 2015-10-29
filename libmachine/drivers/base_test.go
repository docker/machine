package drivers

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIP(t *testing.T) {
	cases := []struct {
		baseDriver  *BaseDriver
		expectedIp  string
		expectedErr error
	}{
		{&BaseDriver{}, "", errors.New("IP address is not set")},
		{&BaseDriver{IPAddress: "2001:4860:0:2001::68"}, "2001:4860:0:2001::68", nil},
		{&BaseDriver{IPAddress: "192.168.0.1"}, "192.168.0.1", nil},
		{&BaseDriver{IPAddress: "::1"}, "::1", nil},
		{&BaseDriver{IPAddress: "whatever"}, "", fmt.Errorf("IP address is invalid: %s", "whatever")},
	}

	for _, c := range cases {
		ip, err := c.baseDriver.GetIP()
		assert.Equal(t, c.expectedIp, ip)
		assert.Equal(t, c.expectedErr, err)
	}
}
