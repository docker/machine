package api

import (
	"testing"

	_ "github.com/docker/machine/drivers/none"
)

func TestStoreList(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	api, err := getTestApi()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := api.Create("test", "none", flags); err != nil {
		t.Fatal(err)
	}

	machines, errs := api.List()
	if len(errs) != 0 {
		t.Fatal(errs)
	}

	if len(machines) != 1 {
		t.Fatalf("list returned %d items", len(machines))
	}

	if machines[0].Name != "test" {
		t.Fatalf("machines[0] name is incorrect, got: %s", machines[0].Name)
	}

	if err := clearHosts(); err != nil {
		t.Fatal(err)
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

	api, err := getTestApi()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := api.Create("test", "none", flags); err != nil {
		t.Fatal(err)
	}

	exists, err := api.Exists("test")
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("exists returned false when it should have been true")
	}

	if err := clearHosts(); err != nil {
		t.Fatal(err)
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

	api, err := getTestApi()
	if err != nil {
		t.Fatal(err)
	}

	machine, err := api.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	api, err = getTestApi()
	if err != nil {
		t.Fatal(err)
	}

	machine, apiErr := api.Get("test")
	if apiErr != nil {
		t.Fatal(apiErr)
	}

	if machine.Name != "test" {
		t.Fatal("Host name is incorrect")
	}

	actualURL, err := machine.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	if actualURL != expectedURL {
		t.Fatalf("GetURL is not %q, got %q", expectedURL, expectedURL)
	}

	if err := clearHosts(); err != nil {
		t.Fatal(err)
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

	api, err := getTestApi()
	if err != nil {
		t.Fatal(err)
	}

	// No hosts set
	machine, err := api.GetActive()
	if err != nil {
		t.Fatal(err)
	}

	if machine != nil {
		t.Fatalf("GetActive: active machine should not exist")
	}

	// Set normal machine
	originalMachine, err := api.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	if err := api.SetActive(originalMachine); err != nil {
		t.Fatal(err)
	}

	machine, err = api.GetActive()
	if err != nil {
		t.Fatal(err)
	}

	if machine.Name != "test" {
		t.Fatalf("active machine is not 'test', got %s", machine.Name)
	}

	isActive, err := api.IsActive(machine)
	if err != nil {
		t.Fatal(err)
	}

	if isActive != true {
		t.Fatal("IsActive: active machine is not test")
	}

	// remove active host altogether
	if err := api.RemoveActive(); err != nil {
		t.Fatal(err)
	}

	machine, err = api.GetActive()
	if err != nil {
		t.Fatal(err)
	}

	if machine != nil {
		t.Fatalf("active machine %s is not nil", machine.Name)
	}

	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}
}
