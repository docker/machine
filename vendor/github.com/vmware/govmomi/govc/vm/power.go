/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

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

package vm

import (
	"flag"
	"fmt"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"golang.org/x/net/context"
)

type power struct {
	*flags.ClientFlag
	*flags.SearchFlag

	On       bool
	Off      bool
	Reset    bool
	Reboot   bool
	Shutdown bool
	Suspend  bool
	Force    bool
}

func init() {
	cli.Register("vm.power", &power{})
}

func (cmd *power) Register(f *flag.FlagSet) {
	cmd.SearchFlag = flags.NewSearchFlag(flags.SearchVirtualMachines)

	f.BoolVar(&cmd.On, "on", false, "Power on")
	f.BoolVar(&cmd.Off, "off", false, "Power off")
	f.BoolVar(&cmd.Reset, "reset", false, "Power reset")
	f.BoolVar(&cmd.Suspend, "suspend", false, "Power suspend")
	f.BoolVar(&cmd.Reboot, "r", false, "Reboot guest")
	f.BoolVar(&cmd.Shutdown, "s", false, "Shutdown guest")
	f.BoolVar(&cmd.Force, "force", false, "Force (ignore state error)")
}

func (cmd *power) Process() error {
	opts := []bool{cmd.On, cmd.Off, cmd.Reset, cmd.Suspend, cmd.Reboot, cmd.Shutdown}
	selected := false

	for _, opt := range opts {
		if opt {
			if selected {
				return flag.ErrHelp
			}
			selected = opt
		}
	}

	if !selected {
		return flag.ErrHelp
	}

	return nil
}

func (cmd *power) Run(f *flag.FlagSet) error {
	vms, err := cmd.VirtualMachines(f.Args())
	if err != nil {
		return err
	}

	for _, vm := range vms {
		var task *object.Task

		switch {
		case cmd.On:
			fmt.Fprintf(cmd, "Powering on %s... ", vm.Reference())
			task, err = vm.PowerOn(context.TODO())
		case cmd.Off:
			fmt.Fprintf(cmd, "Powering off %s... ", vm.Reference())
			task, err = vm.PowerOff(context.TODO())
		case cmd.Reset:
			fmt.Fprintf(cmd, "Reset %s... ", vm.Reference())
			task, err = vm.Reset(context.TODO())
		case cmd.Suspend:
			fmt.Fprintf(cmd, "Suspend %s... ", vm.Reference())
			task, err = vm.Suspend(context.TODO())
		case cmd.Reboot:
			fmt.Fprintf(cmd, "Reboot guest %s... ", vm.Reference())
			err = vm.RebootGuest(context.TODO())
		case cmd.Shutdown:
			fmt.Fprintf(cmd, "Shutdown guest %s... ", vm.Reference())
			err = vm.ShutdownGuest(context.TODO())
		}

		if err != nil {
			return err
		}

		if task != nil {
			err = task.Wait(context.TODO())
		}
		if err == nil {
			fmt.Fprintf(cmd, "OK\n")
			continue
		}

		if cmd.Force {
			fmt.Fprintf(cmd, "Error: %s\n", err)
			continue
		}

		return err
	}

	return nil
}
