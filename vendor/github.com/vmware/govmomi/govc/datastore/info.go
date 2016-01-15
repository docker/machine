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

package datastore

import (
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type info struct {
	*flags.ClientFlag
	*flags.OutputFlag
	*flags.DatacenterFlag
}

func init() {
	cli.Register("datastore.info", &info{})
}

func (cmd *info) Register(f *flag.FlagSet) {}

func (cmd *info) Process() error { return nil }

func (cmd *info) Usage() string {
	return "[PATH]..."
}

func (cmd *info) Run(f *flag.FlagSet) error {
	c, err := cmd.Client()
	if err != nil {
		return err
	}

	ctx := context.TODO()

	finder, err := cmd.Finder()
	if err != nil {
		return err
	}

	args := f.Args()
	if len(args) == 0 {
		args = []string{"*"}
	}

	datastores, err := finder.DatastoreList(ctx, args[0])
	if err != nil {
		return err
	}

	refs := make([]types.ManagedObjectReference, 0, len(datastores))
	for _, ds := range datastores {
		refs = append(refs, ds.Reference())
	}

	var res infoResult
	var props []string

	if cmd.OutputFlag.JSON {
		props = nil // Load everything
	} else {
		props = []string{"info", "summary"} // Load summary
	}

	pc := property.DefaultCollector(c)
	err = pc.Retrieve(ctx, refs, props, &res.Datastores)
	if err != nil {
		return err
	}

	return cmd.WriteResult(&res)
}

type infoResult struct {
	Datastores []mo.Datastore
}

func (r *infoResult) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)

	for _, ds := range r.Datastores {
		s := ds.Summary
		fmt.Fprintf(tw, "Name:\t%s\n", s.Name)
		fmt.Fprintf(tw, "  Type:\t%s\n", s.Type)
		fmt.Fprintf(tw, "  URL:\t%s\n", s.Url)
		fmt.Fprintf(tw, "  Capacity:\t%.1f GB\n", float64(s.Capacity)/(1<<30))
		fmt.Fprintf(tw, "  Free:\t%.1f GB\n", float64(s.FreeSpace)/(1<<30))

		switch info := ds.Info.(type) {
		case *types.NasDatastoreInfo:
			fmt.Fprintf(tw, "  Remote:\t%s:%s\n", info.Nas.RemoteHost, info.Nas.RemotePath)
		}
	}

	return tw.Flush()
}
