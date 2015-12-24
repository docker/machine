/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

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

package host

import (
	"errors"
	"flag"
	"fmt"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type add struct {
	*flags.ClientFlag
	*flags.DatacenterFlag

	parent string

	host        string
	username    string
	password    string
	connect     bool
	fingerprint string
}

func init() {
	cli.Register("host.add", &add{})
}

func (cmd *add) Register(f *flag.FlagSet) {
	f.StringVar(&cmd.parent, "parent", "", "Path to folder to add the host to")

	f.StringVar(&cmd.host, "host", "", "Hostname or IP address of the host")
	f.StringVar(&cmd.username, "username", "", "Username of administration account on the host")
	f.StringVar(&cmd.password, "password", "", "Password of administration account on the host")
	f.BoolVar(&cmd.connect, "connect", true, "Immediately connect to host")
	f.StringVar(&cmd.fingerprint, "fingerprint", "", "Fingerprint of the host's SSL certificate")
}

func (cmd *add) Process() error {
	if cmd.host == "" {
		return flag.ErrHelp
	}
	if cmd.username == "" {
		return flag.ErrHelp
	}
	if cmd.password == "" {
		return flag.ErrHelp
	}
	return nil
}

func (cmd *add) Usage() string {
	return "HOST"
}

func (cmd *add) Description() string {
	return `Add HOST to datacenter.

The host is added to the folder specified by the 'parent' flag. If not given,
this defaults to the hosts folder in the specified or default datacenter.`
}

func (cmd *add) Run(f *flag.FlagSet) error {
	var ctx = context.Background()
	var parent *object.Folder

	client, err := cmd.Client()
	if err != nil {
		return err
	}

	if cmd.parent == "" {
		dc, err := cmd.Datacenter()
		if err != nil {
			return err
		}

		folders, err := dc.Folders(ctx)
		if err != nil {
			return err
		}

		parent = folders.HostFolder
	} else {
		finder, err := cmd.Finder()
		if err != nil {
			return err
		}

		mo, err := finder.ManagedObjectList(ctx, cmd.parent)
		if err != nil {
			return err
		}

		if len(mo) == 0 {
			return errors.New("parent does not resolve to object")
		}

		if len(mo) > 1 {
			return errors.New("parent resolves to more than one object")
		}

		ref := mo[0].Object.Reference()
		if ref.Type != "Folder" {
			return errors.New("parent does not resolve to folder")
		}

		parent = object.NewFolder(client, ref)
	}

	req := types.AddStandaloneHost_Task{
		This: parent.Reference(),
		Spec: types.HostConnectSpec{
			HostName:      cmd.host,
			UserName:      cmd.username,
			Password:      cmd.password,
			SslThumbprint: cmd.fingerprint,
		},
		AddConnected: cmd.connect,
	}

	res, err := methods.AddStandaloneHost_Task(ctx, client, &req)
	if err != nil {
		return err
	}

	task := object.NewTask(client, res.Returnval)
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		f, ok := err.(types.HasFault)
		if !ok {
			return err
		}

		switch fault := f.Fault().(type) {
		case *types.SSLVerifyFault:
			// Add fingerprint to error message
			return fmt.Errorf("%s Fingerprint is %s.", err.Error(), fault.Thumbprint)
		default:
			return err
		}
	}

	return nil
}
