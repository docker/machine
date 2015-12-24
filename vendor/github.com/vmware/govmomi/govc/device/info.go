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
	"os"
	"strings"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type info struct {
	*flags.VirtualMachineFlag
}

func init() {
	cli.Register("device.info", &info{})
}

func (cmd *info) Register(f *flag.FlagSet) {}

func (cmd *info) Process() error { return nil }

func (cmd *info) Usage() string {
	return "DEVICE..."
}

func (cmd *info) Run(f *flag.FlagSet) error {
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

	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)

	for _, name := range f.Args() {
		device := devices.Find(name)
		if device == nil {
			return fmt.Errorf("device '%s' not found", name)
		}

		d := device.GetVirtualDevice()
		info := d.DeviceInfo.GetDescription()

		fmt.Fprintf(tw, "Name:\t%s\n", name)
		fmt.Fprintf(tw, "  Type:\t%s\n", devices.TypeName(device))
		fmt.Fprintf(tw, "  Label:\t%s\n", info.Label)
		fmt.Fprintf(tw, "  Summary:\t%s\n", info.Summary)
		fmt.Fprintf(tw, "  Key:\t%d\n", d.Key)

		if c, ok := device.(types.BaseVirtualController); ok {
			var attached []string
			for _, key := range c.GetVirtualController().Device {
				attached = append(attached, devices.Name(devices.FindByKey(key)))
			}
			fmt.Fprintf(tw, "  Devices:\t%s\n", strings.Join(attached, ", "))
		} else {
			if c := devices.FindByKey(d.ControllerKey); c != nil {
				fmt.Fprintf(tw, "  Controller:\t%s\n", devices.Name(c))
				fmt.Fprintf(tw, "  Unit number:\t%d\n", d.UnitNumber)
			}
		}

		if ca := d.Connectable; ca != nil {
			fmt.Fprintf(tw, "  Connected:\t%t\n", ca.Connected)
			fmt.Fprintf(tw, "  Start connected:\t%t\n", ca.StartConnected)
			fmt.Fprintf(tw, "  Guest control:\t%t\n", ca.AllowGuestControl)
			fmt.Fprintf(tw, "  Status:\t%s\n", ca.Status)
		}

		switch md := device.(type) {
		case types.BaseVirtualEthernetCard:
			fmt.Fprintf(tw, "  MAC Address:\t%s\n", md.GetVirtualEthernetCard().MacAddress)
			fmt.Fprintf(tw, "  Address type:\t%s\n", md.GetVirtualEthernetCard().AddressType)
		case *types.VirtualDisk:
			if b, ok := md.Backing.(types.BaseVirtualDeviceFileBackingInfo); ok {
				fmt.Fprintf(tw, "  File:\t%s\n", b.GetVirtualDeviceFileBackingInfo().FileName)
			}
			if b, ok := md.Backing.(*types.VirtualDiskFlatVer2BackingInfo); ok && b.Parent != nil {
				fmt.Fprintf(tw, "  Parent:\t%s\n", b.Parent.GetVirtualDeviceFileBackingInfo().FileName)
			}
		}
	}

	return tw.Flush()
}
