package google

import (
	"errors"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
)

func TestDriverName(t *testing.T) {
	driverName := newDriver().DriverName()

	assert.Equal(t, "google", driverName)
}

func TestFailWithoutProject(t *testing.T) {
	_, err := create([]string{"test-name"})

	assert.EqualError(t, err, "Please specify the Google Cloud Project name using the option --google-project.")
}

func TestParseFlags(t *testing.T) {
	driver, err := create([]string{
		"--google-project", "test-project",
		"--google-zone", "test-zone",
		"--google-machine-type", "test-machine-type",
		"--google-disk-size", "10000",
		"--google-disk-type", "test-disk-type",
		"--google-address", "test-address",
		"--google-preemptible",
		"--google-auth-token", "test-auth-token-path",
		"--google-scopes", "test-scope1,test-scope2",
		"--google-tags", "test-tag1,test-tag2",
		"--google-username", "test-ssh-user",
		"test-name"})

	assert.NoError(t, err)
	assert.Equal(t, driver.Project, "test-project")
	assert.Equal(t, driver.Zone, "test-zone")
	assert.Equal(t, driver.MachineType, "test-machine-type")
	assert.Equal(t, driver.DiskSize, 10000)
	assert.Equal(t, driver.DiskType, "test-disk-type")
	assert.Equal(t, driver.Address, "test-address")
	assert.Equal(t, driver.Preemptible, true)
	assert.Equal(t, driver.AuthTokenPath, "test-auth-token-path")
	assert.Equal(t, driver.Scopes, "test-scope1,test-scope2")
	assert.Equal(t, driver.Tags, "test-tag1,test-tag2")
	assert.Equal(t, driver.SSHUser, "test-ssh-user")
	assert.Equal(t, driver.SSHPort, 22)
}

func TestSSHHostname(t *testing.T) {
	mockComputeUtilToReturn(&ComputeUtil{ipAddress: "host"})

	hostname, err := newDriver().GetSSHHostname()

	assert.Equal(t, "host", hostname)
	assert.NoError(t, err)
}

func TestGetIp(t *testing.T) {
	mockComputeUtilToReturn(&ComputeUtil{ipAddress: "host"})

	ip, err := newDriver().GetIP()

	assert.Equal(t, "host", ip)
	assert.NoError(t, err)
}

func TestGetIpFailure(t *testing.T) {
	mockComputeUtilToFail("Fail to create ComputeUtil")

	ip, err := newDriver().GetIP()

	assert.Empty(t, ip)
	assert.EqualError(t, err, "Fail to create ComputeUtil")
}

func TestGetURL(t *testing.T) {
	mockComputeUtilToReturn(&ComputeUtil{ipAddress: "host"})

	url, err := newDriver().GetURL()

	assert.Equal(t, "tcp://host:2376", url)
	assert.NoError(t, err)
}

func TestGetURLFailure(t *testing.T) {
	mockComputeUtilToFail("Fail to create ComputeUtil")

	url, err := newDriver().GetURL()

	assert.Empty(t, url)
	assert.EqualError(t, err, "Fail to create ComputeUtil")
}

func TestDefaultSSHUsername(t *testing.T) {
	username := newDriver().GetSSHUsername()

	assert.Equal(t, "docker-user", username)
}

func TestCustomSSHUsername(t *testing.T) {
	driver, _ := create([]string{
		"--google-project", "p",
		"--google-username", "custom-user",
		"name"})

	username := driver.GetSSHUsername()

	assert.Equal(t, "custom-user", username)
}

func newDriver() *Driver {
	return NewDriver("", "")
}

// mockComputeUtilToReturn makes sure calling `newComputeUtil` will always
// return the provided `ComputeUtil`.
func mockComputeUtilToReturn(instance *ComputeUtil) {
	newComputeUtil = func(driver *Driver) (*ComputeUtil, error) {
		return instance, nil
	}
}

// mockComputeUtilToReturn makes sure calling `newComputeUtil` will always
// fail with provided message.
func mockComputeUtilToFail(msg string) {
	newComputeUtil = func(driver *Driver) (*ComputeUtil, error) {
		return nil, errors.New(msg)
	}
}

// create parses the provided args and tries to configure the driver for
// a `create` action.
func create(args []string) (*Driver, error) {
	var err error
	var driver *Driver

	app := cli.NewApp()
	app.Name = "create"
	app.Flags = GetCreateFlags()
	app.Action = func(c *cli.Context) {
		driver = NewDriver(c.Args()[0], "/storepath")
		err = driver.SetConfigFromFlags(c)
	}

	args = append([]string{"create"}, args...)
	app.Run(args)

	return driver, err
}
