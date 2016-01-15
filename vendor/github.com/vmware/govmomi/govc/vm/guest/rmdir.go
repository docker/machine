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

	"github.com/vmware/govmomi/govc/cli"
	"golang.org/x/net/context"
)

type rmdir struct {
	*GuestFlag

	recursive bool
}

func init() {
	cli.Register("guest.rmdir", &rmdir{})
}

func (cmd *rmdir) Register(f *flag.FlagSet) {
	f.BoolVar(&cmd.recursive, "p", false, "Recursive removal")
}

func (cmd *rmdir) Process() error { return nil }

func (cmd *rmdir) Run(f *flag.FlagSet) error {
	m, err := cmd.FileManager()
	if err != nil {
		return err
	}

	return m.DeleteDirectory(context.TODO(), cmd.Auth(), f.Arg(0), cmd.recursive)
}
