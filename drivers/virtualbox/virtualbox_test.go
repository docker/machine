package virtualbox

import (
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestDriverName(t *testing.T) {
	driverName := newTestDriver("default").DriverName()

	assert.Equal(t, "virtualbox", driverName)
}

func TestSSHHostname(t *testing.T) {
	hostname, err := newTestDriver("default").GetSSHHostname()

	assert.Equal(t, "127.0.0.1", hostname)
	assert.NoError(t, err)
}

func TestDefaultSSHUsername(t *testing.T) {
	username := newTestDriver("default").GetSSHUsername()

	assert.Equal(t, "docker", username)
}

type VBoxManagerMock struct {
	VBoxCmdManager
	args   string
	stdOut string
	stdErr string
	err    error
}

func (v *VBoxManagerMock) vbmOutErr(args ...string) (string, string, error) {
	if strings.Join(args, " ") == v.args {
		return v.stdOut, v.stdErr, v.err
	}
	return "", "", errors.New("Invalid args")
}

func TestState(t *testing.T) {
	var tests = []struct {
		stdOut string
		state  state.State
	}{
		{`VMState="running"`, state.Running},
		{`VMState="paused"`, state.Paused},
		{`VMState="saved"`, state.Saved},
		{`VMState="poweroff"`, state.Stopped},
		{`VMState="aborted"`, state.Stopped},
		{`VMState="whatever"`, state.None},
		{`VMState=`, state.None},
	}

	for _, expected := range tests {
		driver := newTestDriver("default")
		driver.VBoxManager = &VBoxManagerMock{
			args:   "showvminfo default --machinereadable",
			stdOut: expected.stdOut,
		}

		machineState, err := driver.GetState()

		assert.NoError(t, err)
		assert.Equal(t, expected.state, machineState)
	}
}

func TestStateErrors(t *testing.T) {
	var tests = []struct {
		stdErr   string
		err      error
		finalErr error
	}{
		{"Could not find a registered machine named 'unknown'", errors.New("Bug"), errors.New("machine does not exist")},
		{"", errors.New("Unexpected error"), errors.New("Unexpected error")},
	}

	for _, expected := range tests {
		driver := newTestDriver("default")
		driver.VBoxManager = &VBoxManagerMock{
			args:   "showvminfo default --machinereadable",
			stdErr: expected.stdErr,
			err:    expected.err,
		}

		machineState, err := driver.GetState()

		assert.Equal(t, err, expected.finalErr)
		assert.Equal(t, state.Error, machineState)
	}
}

func TestGetRandomIPinSubnet(t *testing.T) {
	// test IP 1.2.3.4
	testIP := net.IPv4(byte(1), byte(2), byte(3), byte(4))
	newIP, err := getRandomIPinSubnet(testIP)
	if err != nil {
		t.Fatal(err)
	}

	if testIP.Equal(newIP) {
		t.Fatalf("expected different IP (source %s); received %s", testIP.String(), newIP.String())
	}

	if newIP[0] != testIP[0] {
		t.Fatalf("expected first octet of %d; received %d", testIP[0], newIP[0])
	}

	if newIP[1] != testIP[1] {
		t.Fatalf("expected second octet of %d; received %d", testIP[1], newIP[1])
	}

	if newIP[2] != testIP[2] {
		t.Fatalf("expected third octet of %d; received %d", testIP[2], newIP[2])
	}
}

func TestGetIPErrors(t *testing.T) {
	var tests = []struct {
		stdOut   string
		err      error
		finalErr error
	}{
		{`VMState="poweroff"`, nil, errors.New("Host is not running")},
		{"", errors.New("Unable to get state"), errors.New("Unable to get state")},
	}

	for _, expected := range tests {
		driver := newTestDriver("default")
		driver.VBoxManager = &VBoxManagerMock{
			args:   "showvminfo default --machinereadable",
			stdOut: expected.stdOut,
			err:    expected.err,
		}

		ip, err := driver.GetIP()

		assert.Empty(t, ip)
		assert.Equal(t, err, expected.finalErr)

		url, err := driver.GetURL()

		assert.Empty(t, url)
		assert.Equal(t, err, expected.finalErr)
	}
}

func TestParseValidCIDR(t *testing.T) {
	ip, network, err := parseAndValidateCIDR("192.168.100.1/24")

	assert.Equal(t, "192.168.100.1", ip.String())
	assert.Equal(t, "192.168.100.0", network.IP.String())
	assert.Equal(t, "ffffff00", network.Mask.String())
	assert.NoError(t, err)
}

func TestInvalidCIDR(t *testing.T) {
	ip, network, err := parseAndValidateCIDR("192.168.100.1")

	assert.EqualError(t, err, "invalid CIDR address: 192.168.100.1")
	assert.Nil(t, ip)
	assert.Nil(t, network)
}

func TestInvalidNetworkIpCIDR(t *testing.T) {
	ip, network, err := parseAndValidateCIDR("192.168.100.0/24")

	assert.Equal(t, ErrNetworkAddrCidr, err)
	assert.Nil(t, ip)
	assert.Nil(t, network)
}

func newTestDriver(name string) *Driver {
	return NewDriver(name, "")
}

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
