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
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"golang.org/x/net/context"
)

type ls struct {
	*flags.ClientFlag
}

func init() {
	cli.Register("fields.ls", &ls{})
}

func (cmd *ls) Register(f *flag.FlagSet) {}

func (cmd *ls) Process() error { return nil }

func (cmd *ls) Run(f *flag.FlagSet) error {
	ctx := context.TODO()

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	m, err := object.GetCustomFieldsManager(c)
	if err != nil {
		return err
	}

	field, err := m.Field(ctx)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)

	for _, def := range field {
		fmt.Fprintf(tw, "%d\t%s\n", def.Key, def.Name)
	}

	return tw.Flush()
}
