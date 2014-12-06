package main

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/docker/machine/drivers"
	_ "github.com/docker/machine/drivers/none"
)

type DriverOptionsMock struct {
	Data map[string]interface{}
}

func (d DriverOptionsMock) String(key string) string {
	return d.Data[key].(string)
}

func (d DriverOptionsMock) Int(key string) int {
	return d.Data[key].(int)
}

func (d DriverOptionsMock) Bool(key string) bool {
	return d.Data[key].(bool)
}

func clearHosts() error {
	return os.RemoveAll(path.Join(drivers.GetHomeDir(), ".docker", "hosts"))
}

func TestStoreCreate(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	store := NewStore()

	host, err := store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}
	if host.Name != "test" {
		t.Fatal("Host name is incorrect")
	}
	path := filepath.Join(drivers.GetHomeDir(), ".docker", "hosts", "test")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}
}

func TestStoreRemove(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	store := NewStore()
	_, err := store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(drivers.GetHomeDir(), ".docker", "hosts", "test")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}
	err = store.Remove("test", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Host path still exists after remove: %s", path)
	}
}

func TestStoreList(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	store := NewStore()
	_, err := store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}
	hosts, err := store.List()
	if len(hosts) != 1 {
		t.Fatalf("List returned %d items", len(hosts))
	}
	if hosts[0].Name != "test" {
		t.Fatalf("hosts[0] name is incorrect, got: %s", hosts[0].Name)
	}
}

func TestStoreExists(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	store := NewStore()
	exists, err := store.Exists("test")
	if exists {
		t.Fatal("Exists returned true when it should have been false")
	}
	_, err = store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}
	exists, err = store.Exists("test")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Exists returned false when it should have been true")
	}
}

func TestStoreLoad(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	expectedURL := "unix:///foo/baz"
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": expectedURL,
		},
	}

	store := NewStore()
	_, err := store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	store = NewStore()
	host, err := store.Load("test")
	if host.Name != "test" {
		t.Fatal("Host name is incorrect")
	}
	actualURL, err := host.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	if actualURL != expectedURL {
		t.Fatalf("GetURL is not %q, got %q", expectedURL, expectedURL)
	}
}

func TestStoreGetSetActive(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	store := NewStore()

	// No hosts set
	host, err := store.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if host != nil {
		t.Fatalf("GetActive: Active host should not exist")
	}

	// Set normal host
	originalHost, err := store.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.SetActive(originalHost); err != nil {
		t.Fatal(err)
	}

	host, err = store.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if host.Name != "test" {
		t.Fatalf("Active host is not 'test', got %s", host.Name)
	}
	isActive, err := store.IsActive(host)
	if err != nil {
		t.Fatal(err)
	}
	if isActive != true {
		t.Fatal("IsActive: Active host is not test")
	}

	// remove active host altogether
	if err := store.RemoveActive(); err != nil {
		t.Fatal(err)
	}

	host, err = store.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if host != nil {
		t.Fatalf("Active host is not nil", host.Name)
	}
}

func TestStoreRename(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	store := NewStore()
	_, err := store.Create("test1", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Rename("test1", "test2")
	if err != nil {
		t.Fatal(err)
	}
	exists, err := store.Exists("test2")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Exists returned false when it should have been true")
	}
}

func TestStoreRenameActive(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	// Create test1 & test2
	store := NewStore()
	_, err := store.Create("test1", "none", flags)
	if err != nil {
		t.Fatal(err)
	}
	oldActiveHost, err := store.Create("test2", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	// set test2 as Active
	if err := store.SetActive(oldActiveHost); err != nil {
		t.Fatal(err)
	}

	// Rename test2 to test3
	err = store.Rename("test2", "test3")
	if err != nil {
		t.Fatal(err)
	}
	exists, err := store.Exists("test3")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Exists returned false when it should have been true")
	}

	// Active host should be test3
	newActiveHost, err := store.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if newActiveHost.Name != "test3" {
		t.Fatalf("Active host is not 'test3', got %s", newActiveHost.Name)
	}
}

func TestStoreRenameInactive(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	// Create test1 & test2
	store := NewStore()
	oldActiveHost, err := store.Create("test1", "none", flags)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Create("test2", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	// set test1 as Active
	if err := store.SetActive(oldActiveHost); err != nil {
		t.Fatal(err)
	}

	// Rename test2 to test3
	err = store.Rename("test2", "test3")
	if err != nil {
		t.Fatal(err)
	}
	exists, err := store.Exists("test3")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Exists returned false when it should have been true")
	}

	// Active host should be test1
	newActiveHost, err := store.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if newActiveHost.Name != "test1" {
		t.Fatalf("Active host is not 'test1', got %s", newActiveHost.Name)
	}
}
