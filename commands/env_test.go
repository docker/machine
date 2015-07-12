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

	provider, err := libmachine.New(store)
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

	host, err := provider.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err = provider.Get("test-a")
	if err != nil {
		t.Fatalf("error loading host: %v", err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("config", 0)
	set.Parse([]string{"test-a"})
	c := cli.NewContext(nil, set, set)
	c.App = &cli.App{
		Name: "docker-machine-test",
	}
	cmdEnv(c)

	w.Close()

	out := <-outStr

	// parse the output into a map of envvar:value for easier testing below
	envvars := make(map[string]string)
	for _, e := range strings.Split(strings.TrimSpace(out), "\n") {
		if !strings.HasPrefix(e, "export ") {
			continue
		}
		kv := strings.SplitN(e, "=", 2)
		key, value := kv[0], kv[1]
		envvars[strings.Replace(key, "export ", "", 1)] = value
	}

	testMachineDir := filepath.Join(store.GetPath(), "machines", host.Name)

	expected := map[string]string{
		"DOCKER_TLS_VERIFY":   "\"1\"",
		"DOCKER_CERT_PATH":    fmt.Sprintf("\"%s\"", testMachineDir),
		"DOCKER_HOST":         "\"unix:///var/run/docker.sock\"",
		"DOCKER_MACHINE_NAME": `"test-a"`,
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

	provider, err := libmachine.New(store)
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

	host, err := provider.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err = provider.Get("test-a")
	if err != nil {
		t.Fatalf("error loading host: %v", err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("config", 0)
	set.Parse([]string{"test-a"})
	c := cli.NewContext(nil, set, set)
	c.App = &cli.App{
		Name: "docker-machine-test",
	}
	cmdEnv(c)

	w.Close()

	out := <-outStr

	// parse the output into a map of envvar:value for easier testing below
	envvars := make(map[string]string)
	for _, e := range strings.Split(strings.TrimSuffix(out, ";\n"), ";\n") {
		if !strings.HasPrefix(e, "set -x ") {
			continue
		}
		kv := strings.SplitN(strings.Replace(e, "set -x ", "", 1), " ", 2)
		key, value := kv[0], kv[1]
		envvars[key] = value
	}

	testMachineDir := filepath.Join(store.GetPath(), "machines", host.Name)

	expected := map[string]string{
		"DOCKER_TLS_VERIFY":   "\"1\"",
		"DOCKER_CERT_PATH":    fmt.Sprintf("\"%s\"", testMachineDir),
		"DOCKER_HOST":         "\"unix:///var/run/docker.sock\"",
		"DOCKER_MACHINE_NAME": `"test-a"`,
	}

	for k, v := range envvars {
		if v != expected[k] {
			t.Fatalf("Expected %s == <%s>, but was <%s>", k, expected[k], v)
		}
	}
}

func TestCmdEnvPowerShell(t *testing.T) {
	stdout := os.Stdout
	shell := os.Getenv("SHELL")
	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Setenv("MACHINE_STORAGE_PATH", TestStoreDir)
	os.Setenv("SHELL", "")

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

	provider, err := libmachine.New(store)
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

	host, err := provider.Create("test-a", "none", hostOptions, flags)
	if err != nil {
		t.Fatal(err)
	}

	host, err = provider.Get("test-a")
	if err != nil {
		t.Fatalf("error loading host: %v", err)
	}

	outStr := make(chan string)

	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		outStr <- testOutput.String()
	}()

	set := flag.NewFlagSet("config", 0)
	set.Parse([]string{"test-a"})
	set.String("shell", "powershell", "")
	c := cli.NewContext(nil, set, set)
	c.App = &cli.App{
		Name: "docker-machine-test",
	}
	cmdEnv(c)

	w.Close()

	out := <-outStr

	// parse the output into a map of envvar:value for easier testing below
	envvars := make(map[string]string)
	for _, e := range strings.Split(strings.TrimSpace(out), "\n") {
		if !strings.HasPrefix(e, "$Env") {
			continue
		}
		kv := strings.SplitN(e, " = ", 2)
		key, value := kv[0], kv[1]
		envvars[strings.Replace(key, "$Env:", "", 1)] = value
	}

	testMachineDir := filepath.Join(store.GetPath(), "machines", host.Name)

	expected := map[string]string{
		"DOCKER_TLS_VERIFY":   "\"1\"",
		"DOCKER_CERT_PATH":    fmt.Sprintf("\"%s\"", testMachineDir),
		"DOCKER_HOST":         "\"unix:///var/run/docker.sock\"",
		"DOCKER_MACHINE_NAME": `"test-a"`,
	}

	for k, v := range envvars {
		if v != expected[k] {
			t.Fatalf("Expected %s == <%s>, but was <%s>", k, expected[k], v)
		}
	}
}
