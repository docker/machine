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
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type ComputeResource struct {
	Common

	InventoryPath string
}

func NewComputeResource(c *vim25.Client, ref types.ManagedObjectReference) *ComputeResource {
	return &ComputeResource{
		Common: NewCommon(c, ref),
	}
}

func (c ComputeResource) Hosts(ctx context.Context) ([]types.ManagedObjectReference, error) {
	var cr mo.ComputeResource

	err := c.Properties(ctx, c.Reference(), []string{"host"}, &cr)
	if err != nil {
		return nil, err
	}

	return cr.Host, nil
}

func (c ComputeResource) Datastores(ctx context.Context) ([]*Datastore, error) {
	var cr mo.ComputeResource

	err := c.Properties(ctx, c.Reference(), []string{"datastore"}, &cr)
	if err != nil {
		return nil, err
	}

	var dss []*Datastore
	for _, ref := range cr.Datastore {
		ds := NewDatastore(c.c, ref)
		dss = append(dss, ds)
	}

	return dss, nil
}

func (c ComputeResource) ResourcePool(ctx context.Context) (*ResourcePool, error) {
	var cr mo.ComputeResource

	err := c.Properties(ctx, c.Reference(), []string{"resourcePool"}, &cr)
	if err != nil {
		return nil, err
	}

	return NewResourcePool(c.c, *cr.ResourcePool), nil
}

func (c ComputeResource) Destroy(ctx context.Context) (*Task, error) {
	req := types.Destroy_Task{
		This: c.Reference(),
	}

	res, err := methods.Destroy_Task(ctx, c.c, &req)
	if err != nil {
		return nil, err
	}

	return NewTask(c.c, res.Returnval), nil
}
