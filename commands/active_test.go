package commands

import (
	"os"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

var (
	h = host.Host{
		Name:       "foo",
		DriverName: "fakedriver",
		Driver: &fakedriver.Driver{
			MockState: state.Running,
		},
	}
)

func TestIsActive(t *testing.T) {
	cases := []struct {
		dockerHost string
		state      state.State
		expected   bool
	}{
		{"", state.Running, false},
		{"tcp://5.6.7.8:2376", state.Running, false},
		{"tcp://1.2.3.4:2376", state.Stopped, false},
		{"tcp://1.2.3.4:2376", state.Running, true},
		{"tcp://1.2.3.4:3376", state.Running, true},
	}

	for _, c := range cases {
		os.Unsetenv("DOCKER_HOST")
		if c.dockerHost != "" {
			os.Setenv("DOCKER_HOST", c.dockerHost)
		}
		actual, err := isActive(&h, c.state, "tcp://1.2.3.4:2376")
		assert.Equal(t, c.expected, actual, "IsActive(%s, \"%s\") should return %v, but didn't", c.state, c.dockerHost, c.expected)
		assert.NoError(t, err)
	}
}
