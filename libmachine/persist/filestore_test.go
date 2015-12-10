package persist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/hosttest"
	"github.com/stretchr/testify/assert"
)

func cleanup() {
	os.RemoveAll(os.Getenv("MACHINE_STORAGE_PATH"))
}

func getTestStore() Filestore {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mcndirs.BaseDir = tmpDir

	return Filestore{
		Path:             tmpDir,
		CaCertPath:       filepath.Join(tmpDir, "certs", "ca-cert.pem"),
		CaPrivateKeyPath: filepath.Join(tmpDir, "certs", "ca-key.pem"),
	}
}

func TestStoreSave(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	assert.NoError(t, err)

	err = store.Save(h)
	assert.NoError(t, err)

	path := store.hostPath(h.Name)

	_, err = os.Stat(path)
	assert.NoError(t, err)

	files, _ := ioutil.ReadDir(path)

	assert.Len(t, files, 1)
	assert.Equal(t, "config.json", files[0].Name())
}

func TestStoreSaveOmitRawDriver(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	assert.NoError(t, err)

	err = store.Save(h)
	assert.NoError(t, err)

	configJSONPath := store.hostConfigPath(h.Name)
	f, err := os.Open(configJSONPath)
	assert.NoError(t, err)

	configData, err := ioutil.ReadAll(f)
	assert.NoError(t, err)

	fakeHost := make(map[string]interface{})

	err = json.Unmarshal(configData, &fakeHost)
	assert.NoError(t, err)

	if rawDriver, ok := fakeHost["RawDriver"]; ok {
		t.Fatal("Should not have gotten a value for RawDriver reading host from disk but got one: ", rawDriver)
	}
}

func TestStoreRemove(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	assert.NoError(t, err)

	err = store.Save(h)
	assert.NoError(t, err)

	path := store.hostPath(h.Name)
	_, err = os.Stat(path)
	assert.NoError(t, err)

	err = store.Remove(h.Name)
	assert.NoError(t, err)

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Host path still exists after remove: %s", path)
	}
}

func TestStoreList(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	assert.NoError(t, err)

	err = store.Save(h)
	assert.NoError(t, err)

	hosts, err := store.List()
	assert.Len(t, hosts, 1)
	assert.Equal(t, h.Name, hosts[0])
}

func TestStoreExists(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	assert.NoError(t, err)

	exists, err := store.Exists(h.Name)
	assert.False(t, exists)

	err = store.Save(h)
	assert.NoError(t, err)

	exists, err = store.Exists(h.Name)

	assert.True(t, exists)
	assert.NoError(t, err)

	err = store.Remove(h.Name)
	assert.NoError(t, err)

	exists, err = store.Exists(h.Name)

	assert.False(t, exists)
	assert.NoError(t, err)
}

func TestStoreLoad(t *testing.T) {
	defer cleanup()

	expectedURL := "unix:///foo/baz"
	flags := hosttest.GetTestDriverFlags()
	flags.Data["url"] = expectedURL

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	assert.NoError(t, err)

	err = h.Driver.SetConfigFromFlags(flags)
	assert.NoError(t, err)

	err = store.Save(h)
	assert.NoError(t, err)

	h, err = store.Load(h.Name)
	assert.NoError(t, err)

	rawDataDriver, ok := h.Driver.(*host.RawDataDriver)
	assert.True(t, ok)

	realDriver := none.NewDriver(h.Name, store.Path)

	err = json.Unmarshal(rawDataDriver.Data, &realDriver)
	assert.NoError(t, err)

	h.Driver = realDriver

	actualURL, err := h.URL()

	assert.Equal(t, expectedURL, actualURL)
	assert.NoError(t, err)
}
