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

	"github.com/vmware/govmomi/govc/cli"
	"golang.org/x/net/context"
)

type mktemp struct {
	*GuestFlag

	dir    bool
	prefix string
	suffix string
}

func init() {
	cli.Register("guest.mktemp", &mktemp{})
}

func (cmd *mktemp) Register(f *flag.FlagSet) {
	f.BoolVar(&cmd.dir, "d", false, "Make a directory instead of a file")
	f.StringVar(&cmd.prefix, "t", "", "Prefix")
	f.StringVar(&cmd.suffix, "s", "", "Suffix")
}

func (cmd *mktemp) Process() error { return nil }

func (cmd *mktemp) Run(f *flag.FlagSet) error {
	m, err := cmd.FileManager()
	if err != nil {
		return err
	}

	mk := m.CreateTemporaryFile
	if cmd.dir {
		mk = m.CreateTemporaryDirectory
	}

	name, err := mk(context.TODO(), cmd.Auth(), cmd.prefix, cmd.suffix)
	if err != nil {
		return err
	}

	fmt.Println(name)

	return nil
}
