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

package license

import (
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

type Manager struct {
	reference types.ManagedObjectReference

	c *vim25.Client
}

func NewManager(c *vim25.Client) *Manager {
	m := Manager{
		reference: *c.ServiceContent.LicenseManager,

		c: c,
	}

	return &m
}

func (m Manager) Reference() types.ManagedObjectReference {
	return m.reference
}

func mapToKeyValueSlice(m map[string]string) []types.KeyValue {
	r := make([]types.KeyValue, len(m))
	for k, v := range m {
		r = append(r, types.KeyValue{Key: k, Value: v})
	}
	return r
}

func (m Manager) Add(ctx context.Context, key string, labels map[string]string) (types.LicenseManagerLicenseInfo, error) {
	req := types.AddLicense{
		This:       m.Reference(),
		LicenseKey: key,
		Labels:     mapToKeyValueSlice(labels),
	}

	res, err := methods.AddLicense(ctx, m.c, &req)
	if err != nil {
		return types.LicenseManagerLicenseInfo{}, err
	}

	return res.Returnval, nil
}

func (m Manager) Remove(ctx context.Context, key string) error {
	req := types.RemoveLicense{
		This:       m.Reference(),
		LicenseKey: key,
	}

	_, err := methods.RemoveLicense(ctx, m.c, &req)
	return err
}

func (m Manager) Update(ctx context.Context, key string, labels map[string]string) (types.LicenseManagerLicenseInfo, error) {
	req := types.UpdateLicense{
		This:       m.Reference(),
		LicenseKey: key,
		Labels:     mapToKeyValueSlice(labels),
	}

	res, err := methods.UpdateLicense(ctx, m.c, &req)
	if err != nil {
		return types.LicenseManagerLicenseInfo{}, err
	}

	return res.Returnval, nil
}

func (m Manager) List(ctx context.Context) ([]types.LicenseManagerLicenseInfo, error) {
	var mlm mo.LicenseManager

	err := property.DefaultCollector(m.c).RetrieveOne(ctx, m.Reference(), []string{"licenses"}, &mlm)
	if err != nil {
		return nil, err
	}

	return mlm.Licenses, nil
}
