package softlayer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testStoreDir          = ".store-test"
	machineTestName       = "test-host"
	machineTestCaCert     = "test-cert"
	machineTestPrivateKey = "test-key"
)

type DriverOptionsMock struct {
	Data map[string]interface{}
}

func (d DriverOptionsMock) String(key string) string {
	if value, ok := d.Data[key]; ok {
		return value.(string)
	}
	return ""
}

func (d DriverOptionsMock) StringSlice(key string) []string {
	if value, ok := d.Data[key]; ok {
		return value.([]string)
	}
	return []string{}
}

func (d DriverOptionsMock) Int(key string) int {
	if value, ok := d.Data[key]; ok {
		return value.(int)
	}
	return 0
}

func (d DriverOptionsMock) Bool(key string) bool {
	if value, ok := d.Data[key]; ok {
		return value.(bool)
	}
	return false
}

func cleanup() error {
	return os.RemoveAll(testStoreDir)
}

func getTestStorePath() (string, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		return "", err
	}
	os.Setenv("MACHINE_STORAGE_PATH", tmpDir)
	return tmpDir, nil
}

func getDefaultTestDriverFlags() *DriverOptionsMock {
	return &DriverOptionsMock{
		Data: map[string]interface{}{
			"name":         "test",
			"url":          "unix:///var/run/docker.sock",
			"api-key":      "12345",
			"user":         "abcdefg",
			"api-endpoint": "https://api.softlayer.com/rest/v3",
			"image":        "MY_TEST_IMAGE",
		},
	}
}

func getTestDriver() (*Driver, error) {
	storePath, err := getTestStorePath()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	d, err := NewDriver(machineTestName, storePath, machineTestCaCert, machineTestPrivateKey)
	if err != nil {
		return nil, err
	}
	d.SetConfigFromFlags(getDefaultTestDriverFlags())
	drv := d.(*Driver)
	return drv, nil
}

func TestSetConfigFromFlagsSetsImage(t *testing.T) {
	d, err := getTestDriver()

	if assert.NoError(t, err) {
		assert.Equal(t, "MY_TEST_IMAGE", d.deviceConfig.Image)
	}
}

func TestHostnameDefaultsToMachineName(t *testing.T) {
	d, err := getTestDriver()
	if assert.NoError(t, err) {
		assert.Equal(t, machineTestName, d.deviceConfig.Hostname)
	}
}
