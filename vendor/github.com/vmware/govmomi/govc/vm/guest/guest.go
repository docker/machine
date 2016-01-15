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
	"errors"
	"flag"

	"net/url"

	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/guest"
	"github.com/vmware/govmomi/object"
	"golang.org/x/net/context"
)

type GuestFlag struct {
	*flags.ClientFlag
	*flags.VirtualMachineFlag

	*AuthFlag
}

func (flag *GuestFlag) Register(f *flag.FlagSet) {}

func (flag *GuestFlag) Process() error { return nil }

func (flag *GuestFlag) FileManager() (*guest.FileManager, error) {
	c, err := flag.Client()
	if err != nil {
		return nil, err
	}

	vm, err := flag.VirtualMachine()
	if err != nil {
		return nil, err
	}

	o := guest.NewOperationsManager(c, vm.Reference())
	return o.FileManager(context.TODO())
}

func (flag *GuestFlag) ProcessManager() (*guest.ProcessManager, error) {
	c, err := flag.Client()
	if err != nil {
		return nil, err
	}

	vm, err := flag.VirtualMachine()
	if err != nil {
		return nil, err
	}

	o := guest.NewOperationsManager(c, vm.Reference())
	return o.ProcessManager(context.TODO())
}

func (flag *GuestFlag) ParseURL(urlStr string) (*url.URL, error) {
	c, err := flag.Client()
	if err != nil {
		return nil, err
	}

	return c.Client.ParseURL(urlStr)
}

func (flag *GuestFlag) VirtualMachine() (*object.VirtualMachine, error) {
	vm, err := flag.VirtualMachineFlag.VirtualMachine()
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.New("no vm specified")
	}
	return vm, nil
}
