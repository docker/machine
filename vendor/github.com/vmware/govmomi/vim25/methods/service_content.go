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

package methods

import (
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

var serviceInstance = types.ManagedObjectReference{
	Type:  "ServiceInstance",
	Value: "ServiceInstance",
}

func GetServiceContent(ctx context.Context, r soap.RoundTripper) (types.ServiceContent, error) {
	req := types.RetrieveServiceContent{
		This: serviceInstance,
	}

	res, err := RetrieveServiceContent(ctx, r, &req)
	if err != nil {
		return types.ServiceContent{}, err
	}

	return res.Returnval, nil
}
