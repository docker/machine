package main

import (
	"fmt"
	"os/exec"
	"sync"
	"testing"
)

const (
	MACHINE_NAME = "machine-integration-test-%s"
)

func machineCreate(name string, t *testing.T, wg *sync.WaitGroup) {
	mName := fmt.Sprintf(MACHINE_NAME, name)
	fmt.Printf(" - testing create for %s (%s)\n", name, mName)
	runCmd := exec.Command(machineBinary, "create", "-d", name, mName)
	out, exitCode, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Error(out, err)
	}
	if exitCode != 0 {
		t.Errorf("error creating machine: driver: %s; exit code: %d; output: %s", name, exitCode, out)
	}
	wg.Done()
}

func machineStop(name string, t *testing.T, wg *sync.WaitGroup) {
	mName := fmt.Sprintf(MACHINE_NAME, name)
	fmt.Printf(" - testing stop for %s (%s)\n", name, mName)
	runCmd := exec.Command(machineBinary, "stop", mName)
	out, exitCode, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Error(out, err)
	}
	if exitCode != 0 {
		t.Errorf("error stopping machine: driver: %s; exit code: %d; output: %s", name, exitCode, out)
	}
	wg.Done()
}

func machineStart(name string, t *testing.T, wg *sync.WaitGroup) {
	mName := fmt.Sprintf(MACHINE_NAME, name)
	fmt.Printf(" - testing start for %s (%s)\n", name, mName)
	runCmd := exec.Command(machineBinary, "start", mName)
	out, exitCode, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Error(out, err)
	}
	if exitCode != 0 {
		t.Errorf("error starting machine: driver: %s; exit code: %d; output: %s", name, exitCode, out)
	}
	wg.Done()
}

func machineKill(name string, t *testing.T, wg *sync.WaitGroup) {
	mName := fmt.Sprintf(MACHINE_NAME, name)
	fmt.Printf(" - testing kill for %s (%s)\n", name, mName)
	runCmd := exec.Command(machineBinary, "kill", mName)
	out, exitCode, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Error(out, err)
	}
	if exitCode != 0 {
		t.Errorf("error killing machine: driver: %s; exit code: %d; output: %s", name, exitCode, out)
	}
	wg.Done()
}

func machineRm(name string, t *testing.T, wg *sync.WaitGroup) {
	mName := fmt.Sprintf(MACHINE_NAME, name)
	fmt.Printf(" - testing rm for %s (%s)\n", name, mName)
	runCmd := exec.Command(machineBinary, "rm", "-f", mName)
	out, exitCode, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Error(out, err)
	}
	if exitCode != 0 {
		t.Errorf("error removing machine: driver: %s; exit code: %d; output: %s", name, exitCode, out)
	}
	wg.Done()
}

func TestMachineCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var wg sync.WaitGroup
	for _, d := range machineTestDrivers {
		wg.Add(1)
		go machineCreate(d.name, t, &wg)
	}
	wg.Wait()
}

func TestMachineStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var wg sync.WaitGroup
	for _, d := range machineTestDrivers {
		wg.Add(1)
		go machineStop(d.name, t, &wg)
	}
	wg.Wait()
}

func TestMachineStart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var wg sync.WaitGroup
	for _, d := range machineTestDrivers {
		wg.Add(1)
		go machineStart(d.name, t, &wg)
	}
	wg.Wait()
}

func TestMachineKill(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var wg sync.WaitGroup
	for _, d := range machineTestDrivers {
		wg.Add(1)
		go machineKill(d.name, t, &wg)
	}
	wg.Wait()
}

func TestMachineRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var wg sync.WaitGroup
	for _, d := range machineTestDrivers {
		wg.Add(1)
		go machineRm(d.name, t, &wg)
	}
	wg.Wait()
}
