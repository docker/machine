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

package object

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VirtualApp struct {
	*ResourcePool
}

func NewVirtualApp(c *vim25.Client, ref types.ManagedObjectReference) *VirtualApp {
	return &VirtualApp{
		ResourcePool: NewResourcePool(c, ref),
	}
}

func (p VirtualApp) String() string {
	if p.InventoryPath == "" {
		return p.Common.String()
	}
	return fmt.Sprintf("%v @ %v", p.Common, p.InventoryPath)
}

func (p VirtualApp) Name(ctx context.Context) (string, error) {
	var o mo.VirtualApp

	err := p.Properties(ctx, p.Reference(), []string{"name"}, &o)
	if err != nil {
		return "", err
	}

	return o.Name, nil
}

func (p VirtualApp) CreateChildVM_Task(ctx context.Context, config types.VirtualMachineConfigSpec, host *HostSystem) (*Task, error) {
	req := types.CreateChildVM_Task{
		This:   p.Reference(),
		Config: config,
	}

	if host != nil {
		ref := host.Reference()
		req.Host = &ref
	}

	res, err := methods.CreateChildVM_Task(ctx, p.c, &req)
	if err != nil {
		return nil, err
	}

	return NewTask(p.c, res.Returnval), nil
}

func (p VirtualApp) UpdateVAppConfig(ctx context.Context, spec types.VAppConfigSpec) error {
	req := types.UpdateVAppConfig{
		This: p.Reference(),
		Spec: spec,
	}

	_, err := methods.UpdateVAppConfig(ctx, p.c, &req)
	return err
}

func (p VirtualApp) PowerOnVApp_Task(ctx context.Context) error {
	req := types.PowerOnVApp_Task{
		This: p.Reference(),
	}

	_, err := methods.PowerOnVApp_Task(ctx, p.c, &req)
	return err
}

func (p VirtualApp) PowerOffVApp_Task(ctx context.Context, force bool) error {
	req := types.PowerOffVApp_Task{
		This:  p.Reference(),
		Force: force,
	}

	_, err := methods.PowerOffVApp_Task(ctx, p.c, &req)
	return err
}

func (p VirtualApp) SuspendVApp_Task(ctx context.Context) error {
	req := types.SuspendVApp_Task{
		This: p.Reference(),
	}

	_, err := methods.SuspendVApp_Task(ctx, p.c, &req)
	return err
}
