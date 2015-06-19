// Copyright (C) 2015  Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	//"fmt"
	//"flag"
	//"os"
	"testing"

	"github.com/codegangsta/cli"
)

func checkCommand(t *testing.T, name string, command cli.Command) {
	if command.Name != name {
		t.Errorf("Invalid command name: %s", command.Name)
	}
}

func TestServerSubcommands(t *testing.T) {
	commands := Commands[0].Subcommands
	if len(commands) != 4 {
		t.Errorf("Invalid number of subcommands for server")
	}
	checkCommand(t, "list", commands[0])
	checkCommand(t, "get", commands[1])
	checkCommand(t, "delete", commands[2])
	checkCommand(t, "action", commands[3])
}

func TestUserSubcommands(t *testing.T) {
	commands := Commands[1].Subcommands
	if len(commands) != 1 {
		t.Errorf("Invalid number of subcommands for user")
	}
	checkCommand(t, "get", commands[0])
}

func TestOrganizationsSubcommands(t *testing.T) {
	commands := Commands[2].Subcommands
	if len(commands) != 1 {
		t.Errorf("Invalid number of subcommands for organizations")
	}
	checkCommand(t, "list", commands[0])
}

func TestTokensSubcommands(t *testing.T) {
	commands := Commands[3].Subcommands
	if len(commands) != 5 {
		t.Errorf("Invalid number of subcommands for tokens")
	}
	checkCommand(t, "list", commands[0])
	checkCommand(t, "get", commands[1])
	checkCommand(t, "delete", commands[2])
	checkCommand(t, "create", commands[3])
	checkCommand(t, "update", commands[4])
}

func TestVolumesSubcommands(t *testing.T) {
	commands := Commands[4].Subcommands
	if len(commands) != 4 {
		t.Errorf("Invalid number of subcommands for volumes")
	}
	checkCommand(t, "list", commands[0])
	checkCommand(t, "get", commands[1])
	checkCommand(t, "delete", commands[2])
	checkCommand(t, "create", commands[3])
}

func TestImagesSubcommands(t *testing.T) {
	commands := Commands[5].Subcommands
	if len(commands) != 3 {
		t.Errorf("Invalid number of subcommands for images")
	}
	checkCommand(t, "list", commands[0])
	checkCommand(t, "get", commands[1])
	checkCommand(t, "delete", commands[2])
}
