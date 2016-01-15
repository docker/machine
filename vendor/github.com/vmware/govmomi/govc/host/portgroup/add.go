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

package portgroup

import (
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type add struct {
	*flags.HostSystemFlag

	spec types.HostPortGroupSpec
}

func init() {
	cli.Register("host.portgroup.add", &add{})
}

func (cmd *add) Register(f *flag.FlagSet) {
	f.StringVar(&cmd.spec.VswitchName, "vswitch", "", "vSwitch Name")
	f.IntVar(&cmd.spec.VlanId, "vlan", 0, "VLAN ID")
}

func (cmd *add) Process() error { return nil }

func (cmd *add) Usage() string {
	return "NAME"
}

func (cmd *add) Run(f *flag.FlagSet) error {
	ns, err := cmd.HostNetworkSystem()
	if err != nil {
		return err
	}

	cmd.spec.Name = f.Arg(0)

	return ns.AddPortGroup(context.TODO(), cmd.spec)
}
