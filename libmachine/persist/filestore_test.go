package persist

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/hosttest"
	"github.com/stretchr/testify/assert"
)

const (
	storedDriverURL = "1.2.3.4"
)

type FakePluginDriverFactory struct {
	drivers.Driver
}

func (fpdf *FakePluginDriverFactory) NewPluginDriver(string, []byte) (drivers.Driver, error) {
	return fpdf.Driver, nil
}

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
		PluginDriverFactory: &FakePluginDriverFactory{
			&none.Driver{
				URL: storedDriverURL,
			},
		},
	}
}

func TestStoreSave(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.GetMachinesDir(), h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}
}

func TestStoreRemove(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(store.GetMachinesDir(), h.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}

	err = store.Remove(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Host path still exists after remove: %s", path)
	}
}

func TestStoreList(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	hosts, err := store.List()
	if len(hosts) != 1 {
		t.Fatalf("List returned %d items, expected 1", len(hosts))
	}

	if hosts[0] != h.Name {
		t.Fatalf("hosts[0] name is incorrect, got: %s", hosts[0])
	}
}

func TestStoreExists(t *testing.T) {
	defer cleanup()
	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	exists, err := store.Exists(h.Name)
	if exists {
		t.Fatal("Host should not exist before saving")
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Host should exist after saving")
	}

	if err := store.Remove(h.Name); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("Host should not exist after removing")
	}
}

func TestStoreSaveLoad(t *testing.T) {
	defer cleanup()

	store := getTestStore()

	h, err := hosttest.GetDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(h); err != nil {
		t.Fatal(err)
	}

	h, err = store.Load(h.Name)
	if err != nil {
		t.Fatal(err)
	}

	actualURL, err := h.GetURL()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, storedDriverURL, actualURL)
}
