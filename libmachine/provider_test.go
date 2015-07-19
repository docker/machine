package libmachine

import (
	"os"
	"testing"
)

func getCustomTestProvider(store Store) (*Provider, error) {
	flags := getTestDriverFlags()
	return New(store, flags, nil)
}

func TestProviderGetSetActive(t *testing.T) {
	defer cleanup()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	provider, err := getCustomTestProvider(store)

	// No host set
	host, err := provider.GetActive()
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

	active, err := provider.GetActive()
	if err != nil {
		t.Fatal(err)
	}
	if active.Name != host.Name {
		t.Fatalf("Active host is not '%s', got '%s'", active.Name, host.Name)
	}
}
