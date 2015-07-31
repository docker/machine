// Copyright (C) 2015 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
)

// func checkCommand(t *testing.T, name string, command cli.Command) {
// 	if command.Name != name {
// 		t.Errorf("Invalid command name: %s", command.Name)
// 	}
// }

func TestCLICommands(t *testing.T) {
	app := makeApp()
	if len(app.Commands) != 7 {
		t.Errorf("Invalid CLI number of commands")
	}
}

func checkGlobalArgument(flags []cli.Flag, name string) int {
	for i, flag := range flags {
		fmt.Printf("Flag: %v\n", flag.String())
		if strings.HasPrefix(flag.String(), name) {
			return i
		}
	}
	return -1
}

func TestScalewayUserIDArgument(t *testing.T) {
	app := makeApp()
	if checkGlobalArgument(app.Flags, "--scaleway-userid") == -1 {
		t.Errorf("no userid flag")
	}
}

func TestScalewayTokenArgument(t *testing.T) {
	app := makeApp()
	if checkGlobalArgument(app.Flags, "--scaleway-token") == -1 {
		t.Errorf("No token flag")
	}
}

func TestScalewayOrganizationArgument(t *testing.T) {
	app := makeApp()
	if checkGlobalArgument(app.Flags, "--scaleway-organization") == -1 {
		t.Errorf("No organization flag")
	}
}
