// Copyright (C) 2015  Nicolas Lamirault <nicolas.lamirault@gmail.com>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
