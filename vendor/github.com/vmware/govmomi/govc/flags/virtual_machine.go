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

package flags

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/vmware/govmomi/object"
	"golang.org/x/net/context"
)

type VirtualMachineFlag struct {
	*ClientFlag
	*DatacenterFlag
	*SearchFlag

	register sync.Once
	name     string
	vm       *object.VirtualMachine
}

func (flag *VirtualMachineFlag) Register(f *flag.FlagSet) {
	flag.SearchFlag = NewSearchFlag(SearchVirtualMachines)

	flag.register.Do(func() {
		env := "GOVC_VM"
		value := os.Getenv(env)
		usage := fmt.Sprintf("Virtual machine [%s]", env)
		f.StringVar(&flag.name, "vm", value, usage)
	})
}

func (flag *VirtualMachineFlag) Process() error { return nil }

func (flag *VirtualMachineFlag) VirtualMachine() (*object.VirtualMachine, error) {
	if flag.vm != nil {
		return flag.vm, nil
	}

	// Use search flags if specified.
	if flag.SearchFlag.IsSet() {
		vm, err := flag.SearchFlag.VirtualMachine()
		if err != nil {
			return nil, err
		}

		flag.vm = vm
		return flag.vm, nil
	}

	// Never look for a default virtual machine.
	if flag.name == "" {
		return nil, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	flag.vm, err = finder.VirtualMachine(context.TODO(), flag.name)
	return flag.vm, err
}
