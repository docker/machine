/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package autostart

import (
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/vim25/types"
)

type configure struct {
	*AutostartFlag

	defaults types.AutoStartDefaults
}

func init() {
	cli.Register("host.autostart.configure", &configure{})
}

func (cmd *configure) Register(f *flag.FlagSet) {
	cmd.defaults.Enabled = types.NewBool(false)
	f.BoolVar(cmd.defaults.Enabled, "enabled", false, "")

	f.IntVar(&cmd.defaults.StartDelay, "start-delay", 0, "")
	f.StringVar(&cmd.defaults.StopAction, "stop-action", "", "")
	f.IntVar(&cmd.defaults.StopDelay, "stop-delay", 0, "")

	cmd.defaults.WaitForHeartbeat = types.NewBool(false)
	f.BoolVar(cmd.defaults.WaitForHeartbeat, "wait-for-heartbeat", false, "")
}

func (cmd *configure) Process() error { return nil }

func (cmd *configure) Usage() string {
	return ""
}

func (cmd *configure) Run(f *flag.FlagSet) error {
	// Note: this command cannot DISABLE autostart because the "Enabled" field is
	// marked "omitempty", which means that it is not included when it is false.
	// Also see: https://github.com/vmware/govmomi/issues/240
	return cmd.ReconfigureDefaults(cmd.defaults)
}
