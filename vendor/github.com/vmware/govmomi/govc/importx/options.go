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

package importx

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"

	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type Options struct {
	AllDeploymentOptions []string
	Deployment           string

	AllDiskProvisioningOptions []string
	DiskProvisioning           string

	AllIPAllocationPolicyOptions []string
	IPAllocationPolicy           string

	AllIPProtocolOptions []string
	IPProtocol           string

	PropertyMapping []types.KeyValue

	PowerOn      bool
	InjectOvfEnv bool
	WaitForIP    bool
}

type ImportFlag struct {
	*flags.DatacenterFlag

	folder       string
	optionsFpath string
	Options      Options
}

func (flag *ImportFlag) Register(f *flag.FlagSet) {
	f.StringVar(&flag.folder, "folder", "", "Path to folder to add the vm to")
	f.StringVar(&flag.optionsFpath, "options", "", "Options spec file path for vm deployment")
}

func (flag *ImportFlag) Process() error {
	if len(flag.optionsFpath) > 0 {
		f, err := os.Open(flag.optionsFpath)
		if err != nil {
			return err
		}
		defer f.Close()

		o, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(o, &flag.Options); err != nil {
			return err
		}
	}

	return nil
}

func (flag *ImportFlag) Folder() (*object.Folder, error) {
	if len(flag.folder) == 0 {
		dc, err := flag.Datacenter()
		if err != nil {
			return nil, err
		}
		folders, err := dc.Folders(context.TODO())
		if err != nil {
			return nil, err
		}
		return folders.VmFolder, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	mo, err := finder.ManagedObjectList(context.TODO(), flag.folder)
	if err != nil {
		return nil, err
	}
	if len(mo) == 0 {
		return nil, errors.New("folder argument does not resolve to object")
	}
	if len(mo) > 1 {
		return nil, errors.New("folder argument resolves to more than one object")
	}

	ref := mo[0].Object.Reference()
	if ref.Type != "Folder" {
		return nil, errors.New("folder argument does not resolve to folder")
	}

	c, err := flag.Client()
	if err != nil {
		return nil, err
	}
	return object.NewFolder(c, ref), nil
}
