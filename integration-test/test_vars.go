package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
			MachineDriver{
				name: "virtualbox",
			},
			MachineDriver{
				name: "digitalocean",
			},
		}
	}

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
