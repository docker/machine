package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	drivers "github.com/docker/machine/drivers"
	"github.com/docker/machine/state"
)

type FakeDriver struct {
	MockState state.State
}

func (d *FakeDriver) DriverName() string {
	return "fakedriver"
}

func (d *FakeDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d *FakeDriver) GetURL() (string, error) {
	return "", nil
}

func (d *FakeDriver) GetIP() (string, error) {
	return "", nil
}

func (d *FakeDriver) GetState() (state.State, error) {
	return d.MockState, nil
}

func (d *FakeDriver) PreCreateCheck() error {
	return nil
}

func (d *FakeDriver) Create() error {
	return nil
}

func (d *FakeDriver) Remove() error {
	return nil
}

func (d *FakeDriver) Start() error {
	d.MockState = state.Running
	return nil
}

func (d *FakeDriver) Stop() error {
	d.MockState = state.Stopped
	return nil
}

func (d *FakeDriver) Restart() error {
	return nil
}

func (d *FakeDriver) Kill() error {
	return nil
}

func (d *FakeDriver) Upgrade() error {
	return nil
}

func (d *FakeDriver) StartDocker() error {
	return nil
}

func (d *FakeDriver) StopDocker() error {
	return nil
}

func (d *FakeDriver) GetDockerConfigDir() string {
	return ""
}

func (d *FakeDriver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return &exec.Cmd{}, nil
}

func TestGetHosts(t *testing.T) {
	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}
	os.Setenv("MACHINE_STORAGE_PATH", TestStoreDir)

	flags := getDefaultTestDriverFlags()

	store := NewStore(TestMachineDir, "", "")
	var err error

	_, err = store.Create("test-a", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Create("test-b", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	storeHosts, err := store.List()

	if len(storeHosts) != 2 {
		t.Fatalf("List returned %d items", len(storeHosts))
	}

	set := flag.NewFlagSet("start", 0)
	set.Parse([]string{"test-a", "test-b"})

	globalSet := flag.NewFlagSet("-d", 0)
	globalSet.String("-d", "none", "driver")
	globalSet.String("storage-path", store.Path, "storage path")
	globalSet.String("tls-ca-cert", "", "")
	globalSet.String("tls-ca-key", "", "")

	c := cli.NewContext(nil, set, globalSet)

	hosts, err := getHosts(c)
	if err != nil {
		t.Fatal(err)
	}

	if len(hosts) != 2 {
		t.Fatal("Expected %d hosts, got %d hosts", 2, len(hosts))
	}

	os.Setenv("MACHINE_STORAGE_PATH", "")
}

func TestGetHostState(t *testing.T) {
	storePath, err := ioutil.TempDir("", ".docker")
	if err != nil {
		t.Fatal("Error creating tmp dir:", err)
	}
	hostListItems := make(chan hostListItem)

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	hosts := []Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &FakeDriver{
				MockState: state.Running,
			},
			storePath: storePath,
		},
		{
			Name:       "bar",
			DriverName: "fakedriver",
			Driver: &FakeDriver{
				MockState: state.Stopped,
			},
			storePath: storePath,
		},
		{
			Name:       "baz",
			DriverName: "fakedriver",
			Driver: &FakeDriver{
				MockState: state.Running,
			},
			storePath: storePath,
		},
	}
	expected := map[string]state.State{
		"foo": state.Running,
		"bar": state.Stopped,
		"baz": state.Running,
	}
	items := []hostListItem{}
	for _, host := range hosts {
		go getHostState(host, *store, hostListItems)
	}
	for i := 0; i < len(hosts); i++ {
		items = append(items, <-hostListItems)
	}
	for _, item := range items {
		if expected[item.Name] != item.State {
			t.Fatal("Expected state did not match for item", item)
		}
	}
}

func TestRunActionForeachMachine(t *testing.T) {
	storePath, err := ioutil.TempDir("", ".docker")
	if err != nil {
		t.Fatal("Error creating tmp dir:", err)
	}

	// Assume a bunch of machines in randomly started or
	// stopped states.
	machines := []*Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &FakeDriver{
				MockState: state.Running,
			},
			storePath: storePath,
		},
		{
			Name:       "bar",
			DriverName: "fakedriver",
			Driver: &FakeDriver{
				MockState: state.Stopped,
			},
			storePath: storePath,
		},
		{
			Name: "baz",
			// Ssh, don't tell anyone but this
			// driver only _thinks_ it's named
			// virtualbox...  (to test serial actions)
			// It's actually FakeDriver!
			DriverName: "virtualbox",
			Driver: &FakeDriver{
				MockState: state.Stopped,
			},
			storePath: storePath,
		},
		{
			Name:       "spam",
			DriverName: "virtualbox",
			Driver: &FakeDriver{
				MockState: state.Running,
			},
			storePath: storePath,
		},
		{
			Name:       "eggs",
			DriverName: "fakedriver",
			Driver: &FakeDriver{
				MockState: state.Stopped,
			},
			storePath: storePath,
		},
		{
			Name:       "ham",
			DriverName: "fakedriver",
			Driver: &FakeDriver{
				MockState: state.Running,
			},
			storePath: storePath,
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

func TestCmdConfig(t *testing.T) {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	os.Setenv("MACHINE_STORAGE_PATH", TestStoreDir)

	defer func() {
		os.Setenv("MACHINE_STORAGE_PATH", "")
		os.Stdout = stdout
		w.Close()
	}()

	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := getDefaultTestDriverFlags()

	store := NewStore(TestMachineDir, "", "")
	var err error

	_, err = store.Create("test-a", "none", flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err := store.Load("test-a")
	if err != nil {
		t.Fatalf("error loading host: %v", err)
	}

	if err := store.SetActive(host); err != nil {
		t.Fatalf("error setting active host: %v", err)
	}

	set := flag.NewFlagSet("config", 0)

	testOutput := &bytes.Buffer{}

	go io.Copy(testOutput, r)

	c := cli.NewContext(nil, set, set)

	cmdConfig(c)

	if strings.Contains(testOutput.String(), "-H=unix:///var/run/docker.sock") {
		t.Fatalf("Expect docker host URL")
	}
}
