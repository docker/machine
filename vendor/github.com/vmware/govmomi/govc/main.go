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

package main

import (
	"os"

	"github.com/vmware/govmomi/govc/cli"

	_ "github.com/vmware/govmomi/govc/about"
	_ "github.com/vmware/govmomi/govc/datacenter"
	_ "github.com/vmware/govmomi/govc/datastore"
	_ "github.com/vmware/govmomi/govc/device"
	_ "github.com/vmware/govmomi/govc/device/cdrom"
	_ "github.com/vmware/govmomi/govc/device/floppy"
	_ "github.com/vmware/govmomi/govc/device/scsi"
	_ "github.com/vmware/govmomi/govc/device/serial"
	_ "github.com/vmware/govmomi/govc/events"
	_ "github.com/vmware/govmomi/govc/fields"
	_ "github.com/vmware/govmomi/govc/host"
	_ "github.com/vmware/govmomi/govc/host/autostart"
	_ "github.com/vmware/govmomi/govc/host/esxcli"
	_ "github.com/vmware/govmomi/govc/host/portgroup"
	_ "github.com/vmware/govmomi/govc/host/vswitch"
	_ "github.com/vmware/govmomi/govc/importx"
	_ "github.com/vmware/govmomi/govc/license"
	_ "github.com/vmware/govmomi/govc/ls"
	_ "github.com/vmware/govmomi/govc/pool"
	_ "github.com/vmware/govmomi/govc/version"
	_ "github.com/vmware/govmomi/govc/vm"
	_ "github.com/vmware/govmomi/govc/vm/disk"
	_ "github.com/vmware/govmomi/govc/vm/guest"
	_ "github.com/vmware/govmomi/govc/vm/network"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
