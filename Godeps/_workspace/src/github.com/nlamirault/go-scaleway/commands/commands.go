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
	//"log"
	//"os"

	"github.com/codegangsta/cli"

	"github.com/nlamirault/go-scaleway/api"
)

// Commands is the CLI commands
var Commands = []cli.Command{
	{
		Name: "servers",
		Subcommands: []cli.Command{
			commandListServers,
			commandGetServer,
			commandDeleteServer,
			commandActionServer,
		},
	},
	{
		Name: "users",
		Subcommands: []cli.Command{
			commandGetUser,
		},
	},
	{
		Name: "organizations",
		Subcommands: []cli.Command{
			commandListOrganizations,
		},
	},
	{
		Name: "tokens",
		Subcommands: []cli.Command{
			commandListTokens,
			commandGetToken,
			commandDeleteToken,
			commandCreateToken,
			commandUpdateToken,
		},
	},
	{
		Name: "volumes",
		Subcommands: []cli.Command{
			commandListVolumes,
			commandGetVolume,
			commandDeleteVolume,
			commandCreateVolume,
		},
	},
	{
		Name: "images",
		Subcommands: []cli.Command{
			commandListImages,
			commandGetImage,
			commandDeleteImage,
		},
	},
}

// Flags is the default arguments to the CLI.
var verboseFlag = cli.BoolFlag{
	Name:  "verbose",
	Usage: "Show more output",
}

func getClient(c *cli.Context) *api.ScalewayClient {
	return api.NewClient(
		c.GlobalString("scaleway-token"),
		c.GlobalString("scaleway-userid"),
		c.GlobalString("scaleway-organization"))
}
