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

package pool

import (
	"flag"
	"fmt"
	"path"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type create struct {
	*flags.DatacenterFlag
	*ResourceConfigSpecFlag
}

func init() {
	spec := NewResourceConfigSpecFlag()
	spec.SetAllocation(func(a types.BaseResourceAllocationInfo) {
		ra := a.GetResourceAllocationInfo()
		ra.Shares.Level = types.SharesLevelNormal
		ra.ExpandableReservation = types.NewBool(true)
	})

	cli.Register("pool.create", &create{ResourceConfigSpecFlag: spec})
}

func (cmd *create) Register(f *flag.FlagSet) {}

func (cmd *create) Process() error { return nil }

func (cmd *create) Usage() string {
	return "POOL..."
}

func (cmd *create) Description() string {
	return "Create one or more resource POOLs.\n" + poolCreateHelp
}

func (cmd *create) Run(f *flag.FlagSet) error {
	if f.NArg() == 0 {
		return flag.ErrHelp
	}

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	for _, arg := range f.Args() {
		dir := path.Dir(arg)
		base := path.Base(arg)
		parents, err := finder.ResourcePoolList(context.TODO(), dir)
		if err != nil {
			if _, ok := err.(*find.NotFoundError); ok {
				return fmt.Errorf("cannot create resource pool '%s': parent not found", base)
			}
			return err
		}

		for _, parent := range parents {
			_, err = parent.Create(context.TODO(), base, cmd.ResourceConfigSpec)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
