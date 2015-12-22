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
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

var (
	ErrDatastoreDirNotExist  = errors.New("datastore directory does not exist")
	ErrDatastoreFileNotExist = errors.New("datastore file does not exist")
)

type DatastoreFlag struct {
	*DatacenterFlag

	register sync.Once
	name     string
	ds       *object.Datastore
}

func (flag *DatastoreFlag) Register(f *flag.FlagSet) {
	flag.register.Do(func() {
		env := "GOVC_DATASTORE"
		value := os.Getenv(env)
		usage := fmt.Sprintf("Datastore [%s]", env)
		f.StringVar(&flag.name, "ds", value, usage)
	})
}

func (flag *DatastoreFlag) Process() error { return nil }

func (flag *DatastoreFlag) Datastore() (*object.Datastore, error) {
	if flag.ds != nil {
		return flag.ds, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	if flag.name == "" {
		flag.ds, err = finder.DefaultDatastore(context.TODO())
	} else {
		flag.ds, err = finder.Datastore(context.TODO(), flag.name)
	}

	return flag.ds, err
}

func (flag *DatastoreFlag) DatastorePath(name string) (string, error) {
	ds, err := flag.Datastore()
	if err != nil {
		return "", err
	}

	return ds.Path(name), nil
}

func (flag *DatastoreFlag) DatastoreURL(path string) (*url.URL, error) {
	dc, err := flag.Datacenter()
	if err != nil {
		return nil, err
	}

	ds, err := flag.Datastore()
	if err != nil {
		return nil, err
	}

	u, err := ds.URL(context.TODO(), dc, path)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (flag *DatastoreFlag) Stat(file string) (types.BaseFileInfo, error) {
	ds, err := flag.Datastore()
	if err != nil {
		return nil, err
	}

	b, err := ds.Browser(context.TODO())
	if err != nil {
		return nil, err
	}

	spec := types.HostDatastoreBrowserSearchSpec{
		Details: &types.FileQueryFlags{
			FileType:  true,
			FileOwner: types.NewBool(true), // TODO: omitempty is generated, but seems to be required
		},
		MatchPattern: []string{path.Base(file)},
	}

	dsPath := ds.Path(path.Dir(file))
	task, err := b.SearchDatastore(context.TODO(), dsPath, &spec)
	if err != nil {
		return nil, err
	}

	info, err := task.WaitForResult(context.TODO(), nil)
	if err != nil {
		if info != nil && info.Error != nil {
			_, ok := info.Error.Fault.(*types.FileNotFound)
			if ok {
				// FileNotFound means the base path doesn't exist.
				return nil, ErrDatastoreDirNotExist
			}
		}

		return nil, err
	}

	res := info.Result.(types.HostDatastoreBrowserSearchResults)
	if len(res.File) == 0 {
		// File doesn't exist
		return nil, ErrDatastoreFileNotExist
	}

	return res.File[0], nil
}
