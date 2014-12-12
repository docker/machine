package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/none"
)

func clearHosts() error {
	return os.RemoveAll(filepath.Join(drivers.GetHomeDir(), ".docker/hosts"))
}

func TestStoreCreate(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	store := NewStore()
	url := "unix:///var/run/docker.sock"
	host, err := store.Create("test", "none", &none.CreateFlags{URL: &url})
	if err != nil {
		t.Fatal(err)
	}
	if host.Name != "test" {
		t.Fatal("Host name is incorrect")
	}
	path := filepath.Join(drivers.GetHomeDir(), ".docker/hosts/test")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Host path doesn't exist: %s", path)
	}
}

func TestStoreRemove(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	store := NewStore()
	url := "unix:///var/run/docker.sock"
	_, err := store.Create("test", "none", &none.CreateFlags{URL: &url})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(drivers.GetHomeDir(), ".docker/hosts/test")
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

	store := NewStore()
	url := "unix:///var/run/docker.sock"
	_, err := store.Create("test", "none", &none.CreateFlags{URL: &url})
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

	store := NewStore()
	exists, err := store.Exists("test")
	if exists {
		t.Fatal("Exists returned true when it should have been false")
	}
	url := "unix:///var/run/docker.sock"
	_, err = store.Create("test", "none", &none.CreateFlags{URL: &url})
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

	store := NewStore()
	expectedURL := "unix:///foo/baz"
	_, err := store.Create("test", "none", &none.CreateFlags{URL: &expectedURL})
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
	url := "unix:///var/run/docker.sock"
	originalHost, err := store.Create("test", "none", &none.CreateFlags{URL: &url})
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
