package persist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"strings"

	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/version"
	"github.com/stretchr/testify/assert"
)

const (
	defaultHostName    = "test-host"
	hostTestCaCert     = "test-cert"
	hostTestPrivateKey = "test-key"
)

func testHost() *host.Host {
	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          defaultHostName,
		Driver:        none.NewDriver(defaultHostName, "/tmp/artifacts"),
		DriverName:    "none",
		HostOptions: &host.Options{
			EngineOptions: &engine.Options{},
			SwarmOptions:  &swarm.Options{},
			AuthOptions: &auth.Options{
				CaCertPath:       hostTestCaCert,
				CaPrivateKeyPath: hostTestPrivateKey,
			},
		},
	}
}

func cleanup(store *Filestore) {
	os.RemoveAll(store.Path)
}

func testStore() *Filestore {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return NewFilestore(tmpDir, filepath.Join(tmpDir, "certs", "ca-cert.pem"), filepath.Join(tmpDir, "certs", "ca-key.pem"))
}

func TestStoreSave(t *testing.T) {
	store := testStore()
	defer cleanup(store)

	h := testHost()

	err := store.Save(h)
	assert.NoError(t, err)

	path := store.hostPath(h.Name)

	_, err = os.Stat(path)
	assert.NoError(t, err)

	files, _ := ioutil.ReadDir(path)

	assert.Len(t, files, 1)
	assert.Equal(t, "config.json", files[0].Name())
}

func TestStoreIsCaseInsensitive(t *testing.T) {
	store := testStore()
	defer cleanup(store)

	h := testHost()
	h.Name = "CamelCase"

	err := store.Save(h)
	assert.NoError(t, err)

	path := store.hostPath(h.Name)
	assert.True(t, strings.HasSuffix(path, "/machines/camelcase"))

	exists, err := store.Exists("CamelCase")
	assert.True(t, exists)
	assert.NoError(t, err)

	exists, err = store.Exists("camelcase")
	assert.True(t, exists)
	assert.NoError(t, err)

	loadedHost, err := store.Load("CamelCase")
	assert.Equal(t, "CamelCase", loadedHost.Name)
	assert.NoError(t, err)

	loadedHost, err = store.Load("camelcase")
	assert.Equal(t, "CamelCase", loadedHost.Name)
	assert.NoError(t, err)
}

func TestStoreSaveOmitRawDriver(t *testing.T) {
	store := testStore()
	defer cleanup(store)

	h := testHost()

	err := store.Save(h)
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
	store := testStore()
	defer cleanup(store)

	h := testHost()

	err := store.Save(h)
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
	store := testStore()
	defer cleanup(store)

	h := testHost()

	err := store.Save(h)
	assert.NoError(t, err)

	hosts, err := store.List()
	assert.Len(t, hosts, 1)
	assert.Equal(t, h.Name, hosts[0])
}

func TestStoreExists(t *testing.T) {
	store := testStore()
	defer cleanup(store)

	h := testHost()

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
	store := testStore()
	defer cleanup(store)

	h := testHost()

	expectedURL := "unix:///foo/baz"
	h.Driver.(*none.Driver).URL = expectedURL

	err := store.Save(h)
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
