package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	_ "github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/state"
)

const (
	hostTestName       = "test-host"
	hostTestDriverName = "none"
	hostTestCaCert     = "test-cert"
	hostTestPrivateKey = "test-key"
)

var (
	hostTestStorePath string
	TestStoreDir      string
)

func init() {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	TestStoreDir = tmpDir
}

func clearHosts() error {
	return os.RemoveAll(TestStoreDir)
}

func getTestStore() (libmachine.Store, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hostTestStorePath = tmpDir

	os.Setenv("MACHINE_STORAGE_PATH", tmpDir)

	return libmachine.NewFilestore(tmpDir, hostTestCaCert, hostTestPrivateKey), nil
}

func cleanup() {
	os.RemoveAll(hostTestStorePath)
}

func getTestDriverFlags() *DriverOptionsMock {
	name := hostTestName
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"name":            name,
			"url":             "unix:///var/run/docker.sock",
			"swarm":           false,
			"swarm-host":      "",
			"swarm-master":    false,
			"swarm-discovery": "",
		},
	}
	return flags
}

func getDefaultTestHost() (*libmachine.Host, error) {
	engineOptions := &engine.EngineOptions{}
	swarmOptions := &swarm.SwarmOptions{
		Master:    false,
		Host:      "",
		Discovery: "",
		Address:   "",
	}
	authOptions := &auth.AuthOptions{
		CaCertPath:     hostTestCaCert,
		PrivateKeyPath: hostTestPrivateKey,
	}
	hostOptions := &libmachine.HostOptions{
		EngineOptions: engineOptions,
		SwarmOptions:  swarmOptions,
		AuthOptions:   authOptions,
	}
	host, err := libmachine.NewHost(hostTestName, hostTestDriverName, hostOptions)
	if err != nil {
		return nil, err
	}

	flags := getTestDriverFlags()
	if err := host.Driver.SetConfigFromFlags(flags); err != nil {
		return nil, err
	}

	return host, nil
}

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

func TestRunActionForeachMachine(t *testing.T) {
	storePath, err := ioutil.TempDir("", ".docker")
	if err != nil {
		t.Fatal("Error creating tmp dir:", err)
	}

	// Assume a bunch of machines in randomly started or
	// stopped states.
	machines := []*libmachine.Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			StorePath: storePath,
		},
		{
			Name:       "bar",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Stopped,
			},
			StorePath: storePath,
		},
		{
			Name: "baz",
			// Ssh, don't tell anyone but this
			// driver only _thinks_ it's named
			// virtualbox...  (to test serial actions)
			// It's actually FakeDriver!
			DriverName: "virtualbox",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Stopped,
			},
			StorePath: storePath,
		},
		{
			Name:       "spam",
			DriverName: "virtualbox",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			StorePath: storePath,
		},
		{
			Name:       "eggs",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Stopped,
			},
			StorePath: storePath,
		},
		{
			Name:       "ham",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			StorePath: storePath,
		},
	}

	runActionForeachMachine("start", machines)

	expected := map[string]state.State{
		"foo":  state.Running,
		"bar":  state.Running,
		"baz":  state.Running,
		"spam": state.Running,
		"eggs": state.Running,
		"ham":  state.Running,
	}

	for _, machine := range machines {
		state, _ := machine.Driver.GetState()
		if expected[machine.Name] != state {
			t.Fatalf("Expected machine %s to have state %s, got state %s", machine.Name, state, expected[machine.Name])
		}
	}

	// OK, now let's stop them all!
	expected = map[string]state.State{
		"foo":  state.Stopped,
		"bar":  state.Stopped,
		"baz":  state.Stopped,
		"spam": state.Stopped,
		"eggs": state.Stopped,
		"ham":  state.Stopped,
	}

	runActionForeachMachine("stop", machines)

	for _, machine := range machines {
		state, _ := machine.Driver.GetState()
		if expected[machine.Name] != state {
			t.Fatalf("Expected machine %s to have state %s, got state %s", machine.Name, state, expected[machine.Name])
		}
	}
}
