package libmachine

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	_ "github.com/docker/machine/drivers/none"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/state"

	"github.com/stretchr/testify/assert"
)

const (
	hostTestName       = "test-host"
	hostTestDriverName = "none"
	hostTestCaCert     = "test-cert"
	hostTestPrivateKey = "test-key"
)

var (
	hostTestStorePath string
	stdout            *os.File
)

func init() {
	stdout = os.Stdout
}

func getTestStore() (Store, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	hostTestStorePath = tmpDir

	os.Setenv("MACHINE_STORAGE_PATH", tmpDir)

	return NewFilestore(tmpDir, hostTestCaCert, hostTestPrivateKey), nil
}

func cleanup() {
	os.Stdout = stdout
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

func getDefaultTestHost() (*Host, error) {
	hostOptions := &HostOptions{
		EngineOptions: &engine.EngineOptions{},
		SwarmOptions: &swarm.SwarmOptions{
			Master:    false,
			Host:      "",
			Discovery: "",
			Address:   "",
		},
		AuthOptions: &auth.AuthOptions{
			CaCertPath:     hostTestCaCert,
			PrivateKeyPath: hostTestPrivateKey,
		},
	}
	host, err := NewHost(hostTestName, hostTestDriverName, hostOptions)
	if err != nil {
		return nil, err
	}

	flags := getTestDriverFlags()
	if err := host.Driver.SetConfigFromFlags(flags); err != nil {
		return nil, err
	}

	return host, nil
}

func TestLoadHostDoesNotExist(t *testing.T) {
	_, err := LoadHost("nope-not-here", "/nope/doesnotexist")
	if err == nil {
		t.Fatalf("expected error for non-existent host")
	}
}

func TestLoadHostExists(t *testing.T) {
	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}
	authOptions := host.HostOptions.AuthOptions
	if host.Name != hostTestName {
		t.Fatalf("expected name %s; received %s", hostTestName, host.Name)
	}

	if host.DriverName != hostTestDriverName {
		t.Fatalf("expected driver %s; received %s", hostTestDriverName, host.DriverName)
	}

	if authOptions.CaCertPath != hostTestCaCert {
		t.Fatalf("expected ca cert path %s; received %s", hostTestCaCert, authOptions.CaCertPath)
	}

	if authOptions.PrivateKeyPath != hostTestPrivateKey {
		t.Fatalf("expected key path %s; received %s", hostTestPrivateKey, authOptions.PrivateKeyPath)
	}
}

func TestValidateHostnameValid(t *testing.T) {
	hosts := []string{
		"zomg",
		"test-ing",
		"some.h0st",
	}

	for _, v := range hosts {
		isValid := ValidateHostName(v)
		if !isValid {
			t.Fatalf("Thought a valid hostname was invalid: %s", v)
		}
	}
}

func TestValidateHostnameInvalid(t *testing.T) {
	hosts := []string{
		"zom_g",
		"test$ing",
		"some😄host",
	}

	for _, v := range hosts {
		isValid := ValidateHostName(v)
		if isValid {
			t.Fatalf("Thought an invalid hostname was valid: %s", v)
		}
	}
}

func TestHostOptions(t *testing.T) {
	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	host, err := getDefaultTestHost()
	if err != nil {
		t.Fatal(err)
	}

	if err = store.Save(host); err != nil {
		t.Fatal(err)
	}

	if err := host.SaveConfig(); err != nil {
		t.Fatal(err)
	}

	if err := host.LoadConfig(); err != nil {
		t.Fatal(err)
	}

	// cleanup
	if err := store.Remove(hostTestName, true); err != nil {
		t.Fatal(err)
	}
}

func TestPrintIPEmptyGivenLocalEngine(t *testing.T) {
	defer cleanup()
	host, _ := getDefaultTestHost()

	out, w := captureStdout()

	assert.Nil(t, host.PrintIP())
	w.Close()

	assert.Equal(t, "", strings.TrimSpace(<-out))
}

func TestPrintIPPrintsGivenRemoteEngine(t *testing.T) {
	defer cleanup()
	host, _ := getDefaultTestHost()
	host.Driver = &fakedriver.FakeDriver{}

	out, w := captureStdout()

	assert.Nil(t, host.PrintIP())

	w.Close()

	assert.Equal(t, "1.2.3.4", strings.TrimSpace(<-out))
}

func captureStdout() (chan string, *os.File) {
	r, w, _ := os.Pipe()

	// This is reversed in cleanup()
	os.Stdout = w

	out := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		out <- testOutput.String()
	}()

	return out, w
}

func TestGetHostListItems(t *testing.T) {
	defer cleanup()

	hostListItemsChan := make(chan HostListItem)

	store, err := getTestStore()
	if err != nil {
		t.Fatal(err)
	}

	hosts := []Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &fakedriver.FakeDriver{
				MockState: state.Running,
			},
			StorePath: store.GetPath(),
			HostOptions: &HostOptions{
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
			HostOptions: &HostOptions{
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
			HostOptions: &HostOptions{
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

	items := []HostListItem{}
	for _, host := range hosts {
		go getHostState(host, hostListItemsChan)
	}

	for i := 0; i < len(hosts); i++ {
		items = append(items, <-hostListItemsChan)
	}

	for _, item := range items {
		if expected[item.Name] != item.State {
			t.Fatal("Expected state did not match for item", item)
		}
	}
}
