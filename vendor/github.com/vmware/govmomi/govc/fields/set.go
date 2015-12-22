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

package fields

import (
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/list"
	"github.com/vmware/govmomi/object"
	"golang.org/x/net/context"
)

type set struct {
	*flags.DatacenterFlag
}

func init() {
	cli.Register("fields.set", &set{})
}

func (cmd *set) Register(f *flag.FlagSet) {}

func (cmd *set) Process() error { return nil }

func (cmd *set) Usage() string {
	return "KEY VALUE PATH..."
}

func (cmd *set) Run(f *flag.FlagSet) error {
	if f.NArg() < 3 {
		return flag.ErrHelp
	}

	ctx := context.TODO()

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	m, err := object.GetCustomFieldsManager(c)
	if err != nil {
		return err
	}

	var objs []list.Element

	args := f.Args()

	key, err := m.FindKey(ctx, args[0])
	if err != nil {
		return err
	}

	val := args[1]

	for _, arg := range args[2:] {
		es, err := finder.ManagedObjectList(ctx, arg)
		if err != nil {
			return err
		}

		objs = append(objs, es...)
	}

	for _, ref := range objs {
		err := m.Set(ctx, ref.Object.Reference(), key, val)
		if err != nil {
			return err
		}
	}

	return nil
}
