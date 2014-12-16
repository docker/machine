package main

import (
	"fmt"
	"os"
	"os/exec"
)

type (
	MachineDriver struct {
		name string
	}
)

var (
	machineBinary      = "machine"
	machineTestDrivers []MachineDriver
)

func init() {
	machineTestDrivers = []MachineDriver{
		MachineDriver{
			name: "virtualbox",
		},
		MachineDriver{
			name: "digitalocean",
		},
	}
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
