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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/units"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type ls struct {
	*flags.DatastoreFlag
	*flags.OutputFlag

	long bool
}

func init() {
	cli.Register("datastore.ls", &ls{})
}

func (cmd *ls) Register(f *flag.FlagSet) {
	f.BoolVar(&cmd.long, "l", false, "Long listing format")
}

func (cmd *ls) Process() error { return nil }

func (cmd *ls) Usage() string {
	return "[FILE]..."
}

func (cmd *ls) Run(f *flag.FlagSet) error {
	ds, err := cmd.Datastore()
	if err != nil {
		return err
	}

	b, err := ds.Browser(context.TODO())
	if err != nil {
		return err
	}

	args := f.Args()
	if len(args) == 0 {
		args = []string{""}
	}

	result := &listOutput{
		rs:   make([]types.HostDatastoreBrowserSearchResults, 0),
		long: cmd.long,
	}

	for _, arg := range args {
		spec := types.HostDatastoreBrowserSearchSpec{
			MatchPattern: []string{"*"},
		}

		if cmd.long {
			spec.Details = &types.FileQueryFlags{
				FileType:     true,
				FileSize:     true,
				FileOwner:    types.NewBool(true), // TODO: omitempty is generated, but seems to be required
				Modification: true,
			}
		}

		for i := 0; ; i++ {
			r, err := cmd.ListPath(b, arg, spec)
			if err != nil {
				// Treat the argument as a match pattern if not found as directory
				if i == 0 && types.IsFileNotFound(err) {
					spec.MatchPattern[0] = path.Base(arg)
					arg = path.Dir(arg)
					continue
				}

				return err
			}

			// Treat an empty result against match pattern as file not found
			if i == 1 && len(r.File) == 0 {
				return fmt.Errorf("File %s/%s was not found", r.FolderPath, spec.MatchPattern[0])
			}

			result.add(r)
			break
		}
	}

	return cmd.WriteResult(result)
}

func (cmd *ls) ListPath(b *object.HostDatastoreBrowser, path string, spec types.HostDatastoreBrowserSearchSpec) (types.HostDatastoreBrowserSearchResults, error) {
	var res types.HostDatastoreBrowserSearchResults

	path, err := cmd.DatastorePath(path)
	if err != nil {
		return res, err
	}

	task, err := b.SearchDatastore(context.TODO(), path, &spec)
	if err != nil {
		return res, err
	}

	info, err := task.WaitForResult(context.TODO(), nil)
	if err != nil {
		return res, err
	}

	res = info.Result.(types.HostDatastoreBrowserSearchResults)
	return res, nil
}

type listOutput struct {
	rs   []types.HostDatastoreBrowserSearchResults
	long bool
}

func (o *listOutput) add(r types.HostDatastoreBrowserSearchResults) {
	o.rs = append(o.rs, r)
}

// hasMultiplePaths returns whether or not the slice of search results contains
// results from more than one folder path.
func (o *listOutput) hasMultiplePaths() bool {
	if len(o.rs) == 0 {
		return false
	}

	p := o.rs[0].FolderPath

	// Multiple paths if any entry is not equal to the first one.
	for _, e := range o.rs {
		if e.FolderPath != p {
			return true
		}
	}

	return false
}

func (o *listOutput) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.rs)
}

func (o *listOutput) Write(w io.Writer) error {
	// Only include path header if we're dealing with more than one path.
	includeHeader := false
	if o.hasMultiplePaths() {
		includeHeader = true
	}

	tw := tabwriter.NewWriter(w, 3, 0, 2, ' ', 0)
	for i, r := range o.rs {
		if includeHeader {
			if i > 0 {
				fmt.Fprintf(tw, "\n")
			}
			fmt.Fprintf(tw, "%s:\n", r.FolderPath)
		}
		for _, file := range r.File {
			info := file.GetFileInfo()
			if o.long {
				fmt.Fprintf(tw, "%s\t%s\t%s\n", units.ByteSize(info.FileSize), info.Modification.Format("Mon Jan 2 15:04:05 2006"), info.Path)
			} else {
				fmt.Fprintf(tw, "%s\n", info.Path)
			}
		}
	}
	tw.Flush()
	return nil
}
