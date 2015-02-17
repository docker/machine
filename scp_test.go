package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type ScpFakeDriver struct {
	MockState state.State
}

func (d *ScpFakeDriver) DriverName() string {
	return "fake"
}

func (d *ScpFakeDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func (d *ScpFakeDriver) GetURL() (string, error) {
	return "", nil
}

func (d *ScpFakeDriver) GetIP() (string, error) {
	return "12.34.56.78", nil
}

func (d *ScpFakeDriver) GetState() (state.State, error) {
	return d.MockState, nil
}

func (d *ScpFakeDriver) PreCreateCheck() error {
	return nil
}

func (d *ScpFakeDriver) Create() error {
	return nil
}

func (d *ScpFakeDriver) Remove() error {
	return nil
}

func (d *ScpFakeDriver) Start() error {
	return nil
}

func (d *ScpFakeDriver) Stop() error {
	return nil
}

func (d *ScpFakeDriver) Restart() error {
	return nil
}

func (d *ScpFakeDriver) Kill() error {
	return nil
}

func (d *ScpFakeDriver) Upgrade() error {
	return nil
}

func (d *ScpFakeDriver) StartDocker() error {
	return nil
}

func (d *ScpFakeDriver) StopDocker() error {
	return nil
}

func (d *ScpFakeDriver) GetDockerConfigDir() string {
	return ""
}

func (d *ScpFakeDriver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return &exec.Cmd{}, nil
}

func (d *ScpFakeDriver) GetSSHUsername() string {
	return "root"
}

func (d *ScpFakeDriver) GetSSHKeyPath() string {
	return "/fake/keypath/id_rsa"
}

func bootstrapStore(t *testing.T) (*Store, *Host) {
	storePath, err := ioutil.TempDir("", ".docker")
	if err != nil {
		t.Fatalf("Error creating tmp dir: %s", err)
	}
	store := NewStore(storePath, "", "")

	// TODO: HACK!!! We are checking the name "fake" to
	// bypass the ConfigureAuth() in host.go.
	// We should use a Host interface (which we can mock) instead.
	drivers.Register("fake", &drivers.RegisteredDriver{
		New: func(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
			return &ScpFakeDriver{}, nil
		},
		GetCreateFlags: func() []cli.Flag {
			return nil
		},
	})

	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"url": "unix:///var/run/docker.sock",
		},
	}

	host, err := store.Create("myfunhost", "fake", flags)
	if err != nil {
		t.Fatalf("Unexpected error trying to create fake scp host: %s", err)
	}

	return store, host
}

func TestGetInfoForScpArg(t *testing.T) {
	store, _ := bootstrapStore(t)
	expectedPath := "/tmp/foo"
	host, path, opts, err := getInfoForScpArg("/tmp/foo", *store)
	if err != nil {
		t.Fatalf("Unexpected error in local getInfoForScpArg call: %s", err)
	}
	if path != expectedPath {
		t.Fatalf("Path %s not equal to expected path %s", path, expectedPath)
	}
	if host != nil {
		t.Fatal("host should be nil")
	}
	if opts != nil {
		t.Fatal("opts should be nil")
	}

	host, path, opts, err = getInfoForScpArg("myfunhost:/home/docker/foo", *store)
	if err != nil {
		t.Fatal("Unexpected error in machine-based getInfoForScpArg call: %s", err)
	}
	expectedOpts := []string{
		"-i",
		"/fake/keypath/id_rsa",
	}
	for i := range opts {
		if expectedOpts[i] != opts[i] {
			t.Fatalf("Mismatch in returned opts: %s != %s", expectedOpts[i], opts[i])
		}
	}
	if host.Name != "myfunhost" {
		t.Fatal("Expected host.Name to be myfunhost, got %s", host.Name)
	}
	if path != "/home/docker/foo" {
		t.Fatalf("Expected path to be /home/docker/foo, got %s", path)
	}

	host, path, opts, err = getInfoForScpArg("foo:bar:widget", *store)
	if err != ErrMalformedInput {
		t.Fatalf("Didn't get back an error when we were expecting it for malformed args")
	}
}

func TestGenerateLocationArg(t *testing.T) {
	// local arg
	arg, err := generateLocationArg(nil, "/home/docker/foo")
	if err != nil {
		t.Fatalf("Unexpected error generating location arg for local: %s", err)
	}
	if arg != "/home/docker/foo" {
		t.Fatalf("Expected arg to be /home/docker/foo, was %s", arg)
	}

	// remote arg
	_, host := bootstrapStore(t)
	arg, err = generateLocationArg(host, "/home/docker/foo")
	if err != nil {
		t.Fatalf("Unexpected error generating location arg for remote: %s", err)
	}
	if arg != "root@12.34.56.78:/home/docker/foo" {
		t.Fatalf("Expected arg to be root@12.34.56.78, instead it was %s", arg)
	}
}

func TestGetScpCmd(t *testing.T) {
	// TODO: This is a little "integration-ey".  Perhaps
	// make an ScpDispatcher (name?) interface so that the reliant
	// methods can be mocked.
	expectedArgs := append(
		ssh.BaseSSHOptions,
		"-r",
		"-3",
		"-i",
		"/fake/keypath/id_rsa",
		"/tmp/foo",
		"root@12.34.56.78:/home/docker/foo",
	)
	expectedCmd := exec.Command("scp", expectedArgs...)
	store, _ := bootstrapStore(t)
	cmd, err := GetScpCmd("/tmp/foo", "myfunhost:/home/docker/foo", *store)
	if err != nil {
		t.Fatalf("Unexpected err getting scp command: %s", err)
	}
	correct := reflect.DeepEqual(expectedCmd, cmd)
	if !correct {
		fmt.Println(expectedCmd)
		fmt.Println(cmd)
		t.Fatal("Expected scp cmd structs to be equal but there was mismatch")
	}
}
