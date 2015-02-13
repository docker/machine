package api

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/docker/machine/drivers/none"
)

const (
	TestStoreDir       = ".test-store"
	TestCaCertPath     = ""
	TestPrivateKeyPath = ""
)

func getTestApi() (*Api, error) {
	return NewApi(TestStoreDir, TestCaCertPath, TestPrivateKeyPath)
}

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
	return os.RemoveAll(TestStoreDir)
}

func TestApiCreate(t *testing.T) {
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

	machine, err := api.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	if machine.Name != "test" {
		t.Fatal("machine name is incorrect")
	}

	path := filepath.Join(TestStoreDir, "test")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("machine path doesn't exist: %s", path)
	}

	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}
}

func TestApiRemove(t *testing.T) {
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

	machine, err := api.Create("test", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	if machine.Name != "test" {
		t.Fatal("machine name is incorrect")
	}

	path := filepath.Join(TestStoreDir, "test")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("machine path doesn't exist: %s", path)
	}

	err = api.Remove("test", false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("machine path still exists after remove: %s", path)
	}

	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}
}
