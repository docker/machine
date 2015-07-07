package libmachine

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/docker/machine/drivers/none"
	"github.com/docker/machine/utils"
)

type DriverOptionsMock struct {
	Data map[string]interface{}
}

func (d DriverOptionsMock) String(key string) string {
	return d.Data[key].(string)
}

func (d DriverOptionsMock) StringSlice(key string) []string {
	return d.Data[key].([]string)
}

func (d DriverOptionsMock) Int(key string) int {
	return d.Data[key].(int)
}

func (d DriverOptionsMock) Bool(key string) bool {
	return d.Data[key].(bool)
}

func TestStoreSave(t *testing.T) {
	defer cleanup()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(host); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(utils.GetMachineDir(), host.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}
}

func TestStoreRemove(t *testing.T) {
	defer cleanup()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(host); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(utils.GetMachineDir(), host.Name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}
	err = store.Remove(host.Name, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Host path still exists after remove: %s", path)
	}
}

func TestStoreList(t *testing.T) {
	defer cleanup()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Save(host); err != nil {
		t.Fatal(err)
	}

	hosts, err := store.List()
	if len(hosts) != 1 {
		t.Fatalf("List returned %d items", len(hosts))
	}
	if hosts[0].Name != host.Name {
		t.Fatalf("hosts[0] name is incorrect, got: %s", hosts[0].Name)
	}
}

func TestStoreExists(t *testing.T) {
	defer cleanup()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	exists, err := store.Exists(host.Name)
	if exists {
		t.Fatal("Exists returned true when it should have been false")
	}

	if err := store.Save(host); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(host.Name)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Exists returned false when it should have been true")
	}
	if err := store.Remove(host.Name, true); err != nil {
		t.Fatal(err)
	}

	exists, err = store.Exists(host.Name)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("Exists returned true when it should have been false")
	}
}

func TestStoreLoad(t *testing.T) {
	defer cleanup()

	expectedURL := "unix:///foo/baz"
	flags := getTestDriverFlags()
	flags.Data["url"] = expectedURL

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err := host.Driver.SetConfigFromFlags(flags); err != nil {
		t.Fatal(err)
	}

	if err := store.Save(host); err != nil {
		t.Fatal(err)
	}

	host, err = store.Get(host.Name)
	if host.Name != host.Name {
		t.Fatal("Host name is incorrect")
	}
	actualURL, err := host.GetURL()
	if err != nil {
		t.Fatal(err)
	}

	if actualURL != expectedURL {
		t.Fatalf("GetURL is not %q, got %q", expectedURL, actualURL)
	}
}

func TestStoreGetSetActive(t *testing.T) {
	defer cleanup()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	// No host set
	host, err := store.GetActive()
	if err == nil {
		t.Fatal("Expected an error because there is no active host set")
	}

	if host != nil {
		t.Fatalf("GetActive: Active host should not exist")
	}

	host, err = getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	// Set normal host
	if err := store.Save(host); err != nil {
		t.Fatal(err)
	}

	url, err := host.GetURL()
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("DOCKER_HOST", url)

	host, err = store.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if host.Name != host.Name {
		t.Fatalf("Active host is not 'test', got %s", host.Name)
	}
}
