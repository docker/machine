package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers/fakedriver"
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

func (d DriverOptionsMock) Int(key string) int {
	return d.Data[key].(int)
}

func (d DriverOptionsMock) Bool(key string) bool {
	return d.Data[key].(bool)
}
func TestGetHostState(t *testing.T) {
	defer cleanup()

	hostListItems := make(chan hostListItem)

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	hosts := []libmachine.Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			StorePath: store.GetPath(),
			HostOptions: &libmachine.HostOptions{
				SwarmOptions: &swarm.SwarmOptions{
					Master:    false,
					Address:   "",
					Discovery: "",
				},
			},
		},
		{
			Name:       "bar",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Stopped,
			},
			StorePath: store.GetPath(),
			HostOptions: &libmachine.HostOptions{
				SwarmOptions: &swarm.SwarmOptions{
					Master:    false,
					Address:   "",
					Discovery: "",
				},
			},
		},
		{
			Name:       "baz",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			StorePath: store.GetPath(),
			HostOptions: &libmachine.HostOptions{
				SwarmOptions: &swarm.SwarmOptions{
					Master:    false,
					Address:   "",
					Discovery: "",
				},
			},
		},
	}

	for _, h := range hosts {
		if err := store.Save(&h); err != nil {
			t.Fatal(err)
		}
	}

	expected := map[string]state.State{
		"foo": state.Running,
		"bar": state.Stopped,
		"baz": state.Running,
	}

	items := []hostListItem{}
	for _, host := range hosts {
		go getHostState(host, store, hostListItems)
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

func TestCmdConfig(t *testing.T) {
	defer cleanup()

	stdout := os.Stdout
	r, w, _ := os.Pipe()

	os.Stdout = w

	defer func() {
		os.Stdout = stdout
	}()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	mcn, err := libmachine.New(store)
	if err != nil {
		t.Fatal(err)
	}

	flags := getTestDriverFlags()
	hostOptions := &libmachine.HostOptions{
		EngineOptions: &engine.EngineOptions{},
		SwarmOptions: &swarm.SwarmOptions{
			Master:    false,
			Discovery: "",
			Address:   "",
			Host:      "",
		},
		AuthOptions: &auth.AuthOptions{},
	}

	host, err := mcn.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.SetActive(host); err != nil {
		t.Fatalf("error setting active host: %v", err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("config", 0)
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.String("storage-path", store.GetPath(), "")

	c := cli.NewContext(nil, set, globalSet)

	cmdConfig(c)

	w.Close()

	out := <-outStr

	if !strings.Contains(out, "--tlsverify") {
		t.Fatalf("Expect --tlsverify")
	}

	testMachineDir := filepath.Join(store.GetPath(), "machines", host.Name)

	tlscacert := fmt.Sprintf("--tlscacert=\"%s/ca.pem\"", testMachineDir)
	if !strings.Contains(out, tlscacert) {
		t.Fatalf("Expected to find %s in %s", tlscacert, out)
	}

	tlscert := fmt.Sprintf("--tlscert=\"%s/cert.pem\"", testMachineDir)
	if !strings.Contains(out, tlscert) {
		t.Fatalf("Expected to find %s in %s", tlscert, out)
	}

	tlskey := fmt.Sprintf("--tlskey=\"%s/key.pem\"", testMachineDir)
	if !strings.Contains(out, tlskey) {
		t.Fatalf("Expected to find %s in %s", tlskey, out)
	}

	if !strings.Contains(out, "-H=unix:///var/run/docker.sock") {
		t.Fatalf("Expect docker host URL")
	}
}

func TestCmdEnvBash(t *testing.T) {
	stdout := os.Stdout
	shell := os.Getenv("SHELL")
	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Setenv("MACHINE_STORAGE_PATH", TestStoreDir)
	os.Setenv("SHELL", "/bin/bash")

	defer func() {
		os.Setenv("MACHINE_STORAGE_PATH", "")
		os.Setenv("SHELL", shell)
		os.Stdout = stdout
	}()

	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := getTestDriverFlags()

	store, sErr := getTestStore()
	if sErr != nil {
		t.Fatal(sErr)
	}

	mcn, err := libmachine.New(store)
	if err != nil {
		t.Fatal(err)
	}

	hostOptions := &libmachine.HostOptions{
		EngineOptions: &engine.EngineOptions{},
		SwarmOptions: &swarm.SwarmOptions{
			Master:    false,
			Discovery: "",
			Address:   "",
			Host:      "",
		},
		AuthOptions: &auth.AuthOptions{},
	}

	host, err := mcn.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err = mcn.Get("test-a")
	if err != nil {
		t.Fatalf("error loading host: %v", err)
	}

	if err := mcn.SetActive(host); err != nil {
		t.Fatalf("error setting active host: %v", err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("config", 0)
	c := cli.NewContext(nil, set, set)
	cmdEnv(c)

	w.Close()

	out := <-outStr

	// parse the output into a map of envvar:value for easier testing below
	envvars := make(map[string]string)
	for _, e := range strings.Split(strings.TrimSpace(out), "\n") {
		kv := strings.SplitN(e, "=", 2)
		key, value := kv[0], kv[1]
		envvars[strings.Replace(key, "export ", "", 1)] = value
	}

	testMachineDir := filepath.Join(store.GetPath(), "machines", host.Name)

	expected := map[string]string{
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_CERT_PATH":  fmt.Sprintf("\"%s\"", testMachineDir),
		"DOCKER_HOST":       "unix:///var/run/docker.sock",
	}

	for k, v := range envvars {
		if v != expected[k] {
			t.Fatalf("Expected %s == <%s>, but was <%s>", k, expected[k], v)
		}
	}
}

func TestCmdEnvFish(t *testing.T) {
	stdout := os.Stdout
	shell := os.Getenv("SHELL")
	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Setenv("MACHINE_STORAGE_PATH", TestStoreDir)
	os.Setenv("SHELL", "/bin/fish")

	defer func() {
		os.Setenv("MACHINE_STORAGE_PATH", "")
		os.Setenv("SHELL", shell)
		os.Stdout = stdout
	}()

	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

	flags := getTestDriverFlags()

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	mcn, err := libmachine.New(store)
	if err != nil {
		t.Fatal(err)
	}

	hostOptions := &libmachine.HostOptions{
		EngineOptions: &engine.EngineOptions{},
		SwarmOptions: &swarm.SwarmOptions{
			Master:    false,
			Discovery: "",
			Address:   "",
			Host:      "",
		},
		AuthOptions: &auth.AuthOptions{},
	}

	host, err := mcn.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err = mcn.Get("test-a")
	if err != nil {
		t.Fatalf("error loading host: %v", err)
	}

	if err := mcn.SetActive(host); err != nil {
		t.Fatalf("error setting active host: %v", err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("config", 0)
	c := cli.NewContext(nil, set, set)
	cmdEnv(c)

	w.Close()

	out := <-outStr

	// parse the output into a map of envvar:value for easier testing below
	envvars := make(map[string]string)
	for _, e := range strings.Split(strings.TrimSuffix(out, ";\n"), ";\n") {
		kv := strings.SplitN(strings.Replace(e, "set -x ", "", 1), " ", 2)
		key, value := kv[0], kv[1]
		envvars[key] = value
	}

	testMachineDir := filepath.Join(store.GetPath(), "machines", host.Name)

	expected := map[string]string{
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_CERT_PATH":  fmt.Sprintf("\"%s\"", testMachineDir),
		"DOCKER_HOST":       "unix:///var/run/docker.sock",
	}

	for k, v := range envvars {
		if v != expected[k] {
			t.Fatalf("Expected %s == <%s>, but was <%s>", k, expected[k], v)
		}
	}
}
