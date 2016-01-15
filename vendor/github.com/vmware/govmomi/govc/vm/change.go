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
	"strings"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type extraConfig []types.BaseOptionValue

func (e *extraConfig) String() string {
	return fmt.Sprintf("%v", *e)
}

func (e *extraConfig) Set(v string) error {
	r := strings.Split(v, "=")
	if len(r) < 2 {
		return fmt.Errorf("failed to parse extraConfig: %s", v)
	} else if r[1] == "" {
		return fmt.Errorf("empty value: %s", v)
	}
	*e = append(*e, &types.OptionValue{Key: r[0], Value: r[1]})
	return nil
}

type change struct {
	*flags.VirtualMachineFlag

	types.VirtualMachineConfigSpec
	extraConfig extraConfig
}

func init() {
	cli.Register("vm.change", &change{})
}

func (cmd *change) Register(f *flag.FlagSet) {
	f.Int64Var(&cmd.MemoryMB, "m", 0, "Size in MB of memory")
	f.IntVar(&cmd.NumCPUs, "c", 0, "Number of CPUs")
	f.StringVar(&cmd.GuestId, "g", "", "Guest OS")
	f.StringVar(&cmd.Name, "name", "", "Display name")
	f.Var(&cmd.extraConfig, "e", "ExtraConfig. <key>=<value>")
}

func (cmd *change) Process() error { return nil }

func (cmd *change) Run(f *flag.FlagSet) error {
	vm, err := cmd.VirtualMachine()
	if err != nil {
		return err
	}

	if vm == nil {
		return flag.ErrHelp
	}

	if len(cmd.extraConfig) > 0 {
		cmd.VirtualMachineConfigSpec.ExtraConfig = cmd.extraConfig
	}

	task, err := vm.Reconfigure(context.TODO(), cmd.VirtualMachineConfigSpec)
	if err != nil {
		return err
	}

	return task.Wait(context.TODO())
}
