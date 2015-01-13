package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type (
	MachineDriver struct {
		name string
	}
)

var (
	machineBinary      = "machine"
	machineTestDrivers []MachineDriver
	waitInterval       int
	waitDuration       time.Duration
)

func init() {
	// allow filtering driver tests
	if machineTests := os.Getenv("MACHINE_TESTS"); machineTests != "" {
		tests := strings.Split(machineTests, " ")
		for _, test := range tests {
			mcn := MachineDriver{
				name: test,
			}
			machineTestDrivers = append(machineTestDrivers, mcn)
		}
	} else {
		machineTestDrivers = []MachineDriver{
			{
				name: "virtualbox",
			},
			{
				name: "digitalocean",
			},
		}
	}

	interval := os.Getenv("MACHINE_TEST_DURATION")
	if interval == "" {
		interval = "30"
	}
	wait, err := strconv.Atoi(interval)
	if err != nil {
		fmt.Printf("invalid interval: %s\n", err)
		os.Exit(1)
	}

	waitInterval = wait
	waitDuration = time.Duration(time.Duration(waitInterval) * time.Second)

	// find machine binary
	if machineBin := os.Getenv("MACHINE_BINARY"); machineBin != "" {
		machineBinary = machineBin
	} else {
		whichCmd := exec.Command("which", "machine")
		out, _, err := runCommandWithOutput(whichCmd)
		if err == nil {
			machineBinary = stripTrailingCharacters(out)

		} else {
			fmt.Printf("ERROR: couldn't resolve full path to the Machine binary")
			os.Exit(1)
		}
	}
}
