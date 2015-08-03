package commands

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
)

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

	provider, err := libmachine.New(store)
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

	host, err := provider.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("config", 0)
	set.Parse([]string{"test-a"})
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
