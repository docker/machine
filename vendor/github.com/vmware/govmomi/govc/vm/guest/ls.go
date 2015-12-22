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

package guest

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"golang.org/x/net/context"
)

type ls struct {
	*GuestFlag
}

func init() {
	cli.Register("guest.ls", &ls{})
}

func (cmd *ls) Register(f *flag.FlagSet) {}

func (cmd *ls) Process() error { return nil }

func (cmd *ls) Run(f *flag.FlagSet) error {
	m, err := cmd.FileManager()
	if err != nil {
		return err
	}

	offset := 0
	tw := tabwriter.NewWriter(os.Stdout, 3, 0, 2, ' ', 0)

	for {
		info, err := m.ListFiles(context.TODO(), cmd.Auth(), f.Arg(0), offset, 0, f.Arg(1))
		if err != nil {
			return err
		}

		for _, f := range info.Files {
			attr := f.Attributes.GetGuestFileAttributes() // TODO: GuestPosixFileAttributes
			fmt.Fprintf(tw, "%d\t%s\t%s\n", f.Size, attr.ModificationTime.Format("Mon Jan 2 15:04:05 2006"), f.Path)
		}

		_ = tw.Flush()

		if info.Remaining == 0 {
			break
		}
		offset += len(info.Files)
	}

	return nil
}
