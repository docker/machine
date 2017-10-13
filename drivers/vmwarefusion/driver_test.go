package vmwarefusion

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

var skip = !check(vmrunbin) || !check(vdiskmanbin)

func check(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		log.Printf("%q is missing", path)
		return false
	}

	return true
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

func TestDriver(t *testing.T) {
	if skip {
		t.Skip()
	}

	path, err := ioutil.TempDir("", "vmware-driver-test")
	assert.NoError(t, err)

	defer os.RemoveAll(path)

	driver := NewDriver("default", path)

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err = driver.SetConfigFromFlags(checkFlags)
	assert.NoError(t, err)

	driver.(*Driver).Boot2DockerURL = "https://github.com/boot2docker/boot2docker/releases/download/v17.10.0-ce-rc2/boot2docker.iso"

	err = driver.Create()
	assert.NoError(t, err)

	defer driver.Remove()

	st, err := driver.GetState()
	assert.NoError(t, err)
	assert.Equal(t, state.Running, st)

	ip, err := driver.GetIP()
	assert.NoError(t, err)
	assert.NotZero(t, ip)

	username := driver.GetSSHUsername()
	assert.NotZero(t, username)

	key := driver.GetSSHKeyPath()
	assert.NotZero(t, key)

	port, err := driver.GetSSHPort()
	assert.NoError(t, err)
	assert.NotZero(t, port)

	host, err := driver.GetSSHHostname()
	assert.NoError(t, err)
	assert.NotZero(t, host)

}
