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

	"github.com/vmware/govmomi/vim25/types"
)

type FileAttrFlag struct {
	types.GuestPosixFileAttributes
}

func (flag *FileAttrFlag) Register(f *flag.FlagSet) {
	f.IntVar(&flag.OwnerId, "uid", 0, "User ID")
	f.IntVar(&flag.GroupId, "gid", 0, "Group ID")
	f.Int64Var(&flag.Permissions, "perm", 0, "File permissions")
}

func (flag *FileAttrFlag) Process() error {
	return nil
}

func (flag *FileAttrFlag) Attr() types.BaseGuestFileAttributes {
	return &flag.GuestPosixFileAttributes
}
