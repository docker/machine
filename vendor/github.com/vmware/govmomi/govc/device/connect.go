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

package device

import (
	"flag"
	"fmt"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"golang.org/x/net/context"
)

type connect struct {
	*flags.VirtualMachineFlag
}

func init() {
	cli.Register("device.connect", &connect{})
}

func (cmd *connect) Register(f *flag.FlagSet) {}

func (cmd *connect) Process() error { return nil }

func (cmd *connect) Usage() string {
	return "DEVICE..."
}

func (cmd *connect) Run(f *flag.FlagSet) error {
	vm, err := cmd.VirtualMachine()
	if err != nil {
		return err
	}

	if vm == nil {
		return flag.ErrHelp
	}

	devices, err := vm.Device(context.TODO())
	if err != nil {
		return err
	}

	for _, name := range f.Args() {
		device := devices.Find(name)
		if device == nil {
			return fmt.Errorf("device '%s' not found", name)
		}

		if err = devices.Connect(device); err != nil {
			return err
		}

		if err = vm.EditDevice(context.TODO(), device); err != nil {
			return err
		}
	}

	return nil
}
