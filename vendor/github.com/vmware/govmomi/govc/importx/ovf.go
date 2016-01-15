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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type ovfx struct {
	*ImportFlag
	*flags.DatastoreFlag
	*flags.HostSystemFlag
	*flags.OutputFlag
	*flags.ResourcePoolFlag

	Client       *vim25.Client
	Datacenter   *object.Datacenter
	Datastore    *object.Datastore
	ResourcePool *object.ResourcePool

	Archive
}

func init() {
	cli.Register("import.ovf", &ovfx{})
}

func (cmd *ovfx) Register(f *flag.FlagSet) {}

func (cmd *ovfx) Process() error { return nil }

func (cmd *ovfx) Usage() string {
	return "PATH_TO_OVF"
}

func (cmd *ovfx) Run(f *flag.FlagSet) error {
	fpath, err := cmd.Prepare(f)
	if err != nil {
		return err
	}

	cmd.Archive = &FileArchive{fpath}

	moref, err := cmd.Import(fpath)
	if err != nil {
		return err
	}

	vm := object.NewVirtualMachine(cmd.Client, *moref)

	return cmd.Deploy(vm)
}

func (cmd *ovfx) Prepare(f *flag.FlagSet) (string, error) {
	var err error

	args := f.Args()
	if len(args) != 1 {
		return "", errors.New("no file specified")
	}

	cmd.Client, err = cmd.DatastoreFlag.Client()
	if err != nil {
		return "", err
	}

	cmd.Datacenter, err = cmd.DatastoreFlag.Datacenter()
	if err != nil {
		return "", err
	}

	cmd.Datastore, err = cmd.DatastoreFlag.Datastore()
	if err != nil {
		return "", err
	}

	cmd.ResourcePool, err = cmd.ResourcePoolFlag.ResourcePool()
	if err != nil {
		return "", err
	}

	return f.Arg(0), nil
}

func (cmd *ovfx) Deploy(vm *object.VirtualMachine) error {
	if err := cmd.PowerOn(vm); err != nil {
		return err
	}

	if err := cmd.InjectOvfEnv(vm); err != nil {
		return err
	}

	if err := cmd.WaitForIP(vm); err != nil {
		return err
	}

	return nil
}

func (cmd *ovfx) ReadOvf(fpath string) ([]byte, error) {
	f, _, err := cmd.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

func (cmd *ovfx) ReadEnvelope(fpath string) (*ovf.Envelope, error) {
	if fpath == "" {
		return nil, nil
	}

	f, _, err := cmd.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	e, err := ovf.Unmarshal(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ovf: %s", err.Error())
	}

	return e, nil
}

func (cmd *ovfx) Import(fpath string) (*types.ManagedObjectReference, error) {
	o, err := cmd.ReadOvf(fpath)
	if err != nil {
		return nil, err
	}

	e, err := cmd.ReadEnvelope(fpath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ovf: %s", err.Error())
	}

	name := "Govc Virtual Appliance"
	if e.VirtualSystem != nil {
		name = e.VirtualSystem.ID
	}

	cisp := types.OvfCreateImportSpecParams{
		DiskProvisioning:   cmd.Options.DiskProvisioning,
		EntityName:         name,
		IpAllocationPolicy: cmd.Options.IPAllocationPolicy,
		IpProtocol:         cmd.Options.IPProtocol,
		OvfManagerCommonParams: types.OvfManagerCommonParams{
			DeploymentOption: cmd.Options.Deployment,
			Locale:           "US"},
		PropertyMapping: cmd.Options.PropertyMapping,
	}

	m := object.NewOvfManager(cmd.Client)
	spec, err := m.CreateImportSpec(context.TODO(), string(o), cmd.ResourcePool, cmd.Datastore, cisp)
	if err != nil {
		return nil, err
	}
	if spec.Error != nil {
		return nil, errors.New(spec.Error[0].LocalizedMessage)
	}
	if spec.Warning != nil {
		for _, w := range spec.Warning {
			_, _ = cmd.Log(fmt.Sprintf("Warning: %s\n", w.LocalizedMessage))
		}
	}

	// TODO: ImportSpec may have unitNumber==0, but this field is optional in the wsdl
	// and hence omitempty in the struct tag; but unitNumber is required for certain devices.
	s := &spec.ImportSpec.(*types.VirtualMachineImportSpec).ConfigSpec
	for _, d := range s.DeviceChange {
		n := &d.GetVirtualDeviceConfigSpec().Device.GetVirtualDevice().UnitNumber
		if *n == 0 {
			*n = -1
		}
	}

	var host *object.HostSystem
	if cmd.SearchFlag.IsSet() {
		if host, err = cmd.HostSystem(); err != nil {
			return nil, err
		}
	}

	folder, err := cmd.Folder()
	if err != nil {
		return nil, err
	}

	lease, err := cmd.ResourcePool.ImportVApp(context.TODO(), spec.ImportSpec, folder, host)
	if err != nil {
		return nil, err
	}

	info, err := lease.Wait(context.TODO())
	if err != nil {
		return nil, err
	}

	// Build slice of items and URLs first, so that the lease updater can know
	// about every item that needs to be uploaded, and thereby infer progress.
	var items []ovfFileItem

	for _, device := range info.DeviceUrl {
		for _, item := range spec.FileItem {
			if device.ImportKey != item.DeviceId {
				continue
			}

			u, err := cmd.Client.ParseURL(device.Url)
			if err != nil {
				return nil, err
			}

			i := ovfFileItem{
				url:  u,
				item: item,
				ch:   make(chan progress.Report),
			}

			items = append(items, i)
		}
	}

	u := newLeaseUpdater(cmd.Client, lease, items)
	defer u.Done()

	for _, i := range items {
		err = cmd.Upload(lease, i)
		if err != nil {
			return nil, err
		}
	}

	return &info.Entity, lease.HttpNfcLeaseComplete(context.TODO())
}

func (cmd *ovfx) Upload(lease *object.HttpNfcLease, ofi ovfFileItem) error {
	item := ofi.item
	file := item.Path

	f, size, err := cmd.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	logger := cmd.ProgressLogger(fmt.Sprintf("Uploading %s... ", path.Base(file)))
	defer logger.Wait()

	opts := soap.Upload{
		ContentLength: size,
		Progress:      progress.Tee(ofi, logger),
	}

	// Non-disk files (such as .iso) use the PUT method.
	// Overwrite: t header is also required in this case (ovftool does the same)
	if item.Create {
		opts.Method = "PUT"
		opts.Headers = map[string]string{
			"Overwrite": "t",
		}
	} else {
		opts.Method = "POST"
		opts.Type = "application/x-vnd.vmware-streamVmdk"
	}

	return cmd.Client.Client.Upload(f, ofi.url, &opts)
}

func (cmd *ovfx) PowerOn(vm *object.VirtualMachine) error {
	if !cmd.Options.PowerOn {
		return nil
	}

	cmd.Log("Powering on vm...\n")

	task, err := vm.PowerOn(context.TODO())
	if err != nil {
		return err
	}

	if _, err = task.WaitForResult(context.TODO(), nil); err != nil {
		return err
	}

	return nil
}

func (cmd *ovfx) InjectOvfEnv(vm *object.VirtualMachine) error {
	if !cmd.Options.PowerOn || !cmd.Options.InjectOvfEnv {
		return nil
	}

	a := cmd.Client.ServiceContent.About
	if strings.EqualFold(a.ProductLineId, "esx") || strings.EqualFold(a.ProductLineId, "embeddedEsx") {
		cmd.Log("Injecting ovf env...\n")

		// build up Environment in order to marshal to xml
		var epa []ovf.EnvProperty
		for _, p := range cmd.Options.PropertyMapping {
			epa = append(epa, ovf.EnvProperty{
				Key:   p.Key,
				Value: p.Value})
		}
		env := ovf.Env{
			EsxID: vm.Reference().Value,
			Platform: &ovf.PlatformSection{
				Kind:    a.Name,
				Version: a.Version,
				Vendor:  a.Vendor,
				Locale:  "US",
			},
			Property: &ovf.PropertySection{
				Properties: epa},
		}

		xenv := ovf.MarshalManual(env)
		vmConfigSpec := types.VirtualMachineConfigSpec{
			ExtraConfig: []types.BaseOptionValue{&types.OptionValue{
				Key:   "guestinfo.ovfEnv",
				Value: xenv}}}

		task, err := vm.Reconfigure(context.TODO(), vmConfigSpec)
		if err != nil {
			return err
		}
		if err := task.Wait(context.TODO()); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *ovfx) WaitForIP(vm *object.VirtualMachine) error {
	if !cmd.Options.PowerOn || !cmd.Options.WaitForIP {
		return nil
	}

	cmd.Log("Waiting for ip...\n")

	ip, err := vm.WaitForIP(context.TODO())
	if err != nil {
		return err
	}
	cmd.Log(fmt.Sprintf("Received IP address: %s\n", ip))

	return nil
}
