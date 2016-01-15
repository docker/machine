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

package cdrom

import (
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"golang.org/x/net/context"
)

type insert struct {
	*flags.DatastoreFlag
	*flags.VirtualMachineFlag

	device string
}

func init() {
	cli.Register("device.cdrom.insert", &insert{})
}

func (cmd *insert) Register(f *flag.FlagSet) {
	f.StringVar(&cmd.device, "device", "", "CD-ROM device name")
}

func (cmd *insert) Process() error { return nil }

func (cmd *insert) Usage() string {
	return "ISO"
}

func (cmd *insert) Description() string {
	return `Insert media on datastore into CD-ROM device.

If device is not specified, the first CD-ROM device is used.`
}

func (cmd *insert) Run(f *flag.FlagSet) error {
	vm, err := cmd.VirtualMachine()
	if err != nil {
		return err
	}

	if vm == nil || f.NArg() != 1 {
		return flag.ErrHelp
	}

	devices, err := vm.Device(context.TODO())
	if err != nil {
		return err
	}

	c, err := devices.FindCdrom(cmd.device)
	if err != nil {
		return err
	}

	iso, err := cmd.DatastorePath(f.Arg(0))
	if err != nil {
		return nil
	}

	return vm.EditDevice(context.TODO(), devices.InsertIso(c, iso))
}
