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

package vswitch

import (
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"golang.org/x/net/context"
)

type remove struct {
	*flags.HostSystemFlag
}

func init() {
	cli.Register("host.vswitch.remove", &remove{})
}

func (cmd *remove) Register(f *flag.FlagSet) {}

func (cmd *remove) Process() error { return nil }

func (cmd *remove) Usage() string {
	return "NAME"
}

func (cmd *remove) Run(f *flag.FlagSet) error {
	ns, err := cmd.HostNetworkSystem()
	if err != nil {
		return err
	}

	return ns.RemoveVirtualSwitch(context.TODO(), f.Arg(0))
}
