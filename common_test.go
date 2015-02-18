package machine

import (
	"os"
	"path/filepath"
)

var (
	TestStoreDir       = ".store-test"
	TestCaCertPath     = ""
	TestPrivateKeyPath = ""
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

func cleanup() error {
	return os.RemoveAll(TestStoreDir)
}

func getTestMachine() (*Machine, error) {
	name := "test"
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	machinePath := filepath.Join(TestStoreDir, name)

	m, err := NewMachine(name, "none", machinePath, "", "")
	if err != nil {
		return nil, err
	}

	if err := m.Driver.SetConfigFromFlags(flags); err != nil {
		return nil, err
	}

	return m, nil
}
