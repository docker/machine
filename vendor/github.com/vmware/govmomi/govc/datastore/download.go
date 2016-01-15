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
	"errors"
	"flag"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/soap"
)

type download struct {
	*flags.DatastoreFlag
}

func init() {
	cli.Register("datastore.download", &download{})
}

func (cmd *download) Register(f *flag.FlagSet) {}

func (cmd *download) Process() error { return nil }

func (cmd *download) Usage() string {
	return "REMOTE LOCAL"
}

func (cmd *download) Run(f *flag.FlagSet) error {
	args := f.Args()
	if len(args) != 2 {
		return errors.New("invalid arguments")
	}

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	u, err := cmd.DatastoreURL(args[0])
	if err != nil {
		return err
	}

	p := soap.DefaultDownload
	if cmd.OutputFlag.TTY {
		logger := cmd.ProgressLogger("Downloading... ")
		p.Progress = logger
		defer logger.Wait()
	}

	return c.Client.DownloadFile(args[1], u, &p)
}
