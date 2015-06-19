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
	{
		Name: "ips",
		Subcommands: []cli.Command{
			commandListIPs,
			commandGetIP,
			commandDeleteIP,
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
