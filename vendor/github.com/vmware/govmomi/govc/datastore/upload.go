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

type upload struct {
	*flags.OutputFlag
	*flags.DatastoreFlag
}

func init() {
	cli.Register("datastore.upload", &upload{})
}

func (cmd *upload) Register(f *flag.FlagSet) {}

func (cmd *upload) Process() error { return nil }

func (cmd *upload) Usage() string {
	return "LOCAL REMOTE"
}

func (cmd *upload) Run(f *flag.FlagSet) error {
	args := f.Args()
	if len(args) != 2 {
		return errors.New("invalid arguments")
	}

	c, err := cmd.Client()
	if err != nil {
		return err
	}

	u, err := cmd.DatastoreURL(args[1])
	if err != nil {
		return err
	}

	p := soap.DefaultUpload
	if cmd.OutputFlag.TTY {
		logger := cmd.ProgressLogger("Uploading... ")
		p.Progress = logger
		defer logger.Wait()
	}

	return c.Client.UploadFile(args[0], u, &p)
}
