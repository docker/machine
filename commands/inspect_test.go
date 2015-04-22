package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/stretchr/testify/assert"
)

func TestCmdInspectFormat(t *testing.T) {
	actual, host := runInspectCommand(t, []string{})
	expected, _ := json.MarshalIndent(host, "", "    ")
	assert.Equal(t, string(expected), actual)

	actual, _ = runInspectCommand(t, []string{"--format", "{{.DriverName}}"})
	assert.Equal(t, "none", actual)

	actual, _ = runInspectCommand(t, []string{"--format", "{{json .DriverName}}"})
	assert.Equal(t, "\"none\"", actual)

	actual, _ = runInspectCommand(t, []string{"--format", "{{prettyjson .Driver}}"})
	assert.Equal(t, "{\n    \"URL\": \"unix:///var/run/docker.sock\"\n}", actual)
}

func runInspectCommand(t *testing.T, args []string) (string, *libmachine.Host) {
	stdout := os.Stdout
	stderr := os.Stderr
	shell := os.Getenv("SHELL")
	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Stderr = w
	os.Setenv("MACHINE_STORAGE_PATH", TestStoreDir)
	os.Setenv("SHELL", "/bin/bash")

	defer func() {
		os.Setenv("MACHINE_STORAGE_PATH", "")
		os.Setenv("SHELL", shell)
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	if err := clearHosts(); err != nil {
		t.Fatal(err)
	}

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

	flags := getTestDriverFlags()
	_, err = mcn.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("inspect", 0)
	set.String("format", "", "")
	set.Parse(args)
	c := cli.NewContext(nil, set, set)
	cmdInspect(c)

	w.Close()

	out := <-outStr

	return strings.TrimSpace(out), getHost(c)
}
