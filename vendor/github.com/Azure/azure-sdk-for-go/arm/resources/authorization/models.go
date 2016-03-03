package authorization

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator 0.14.0.0
// Changes may cause incorrect behavior and will be lost if the code is
// regenerated.

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"net/http"
)

// LockLevel enumerates the values for lock level.
type LockLevel string

const (
	// CanNotDelete specifies the can not delete state for lock level.
	CanNotDelete LockLevel = "CanNotDelete"
	// NotSpecified specifies the not specified state for lock level.
	NotSpecified LockLevel = "NotSpecified"
	// ReadOnly specifies the read only state for lock level.
	ReadOnly LockLevel = "ReadOnly"
)

// DeploymentExtendedFilter is deployment filter.
type DeploymentExtendedFilter struct {
	ProvisioningState *string `json:"provisioningState,omitempty"`
}

// GenericResourceFilter is resource filter.
type GenericResourceFilter struct {
	ResourceType *string `json:"resourceType,omitempty"`
	Tagname      *string `json:"tagname,omitempty"`
	Tagvalue     *string `json:"tagvalue,omitempty"`
}

// ManagementLockListResult is list of management locks.
type ManagementLockListResult struct {
	autorest.Response `json:"-"`
	Value             *[]ManagementLockObject `json:"value,omitempty"`
	NextLink          *string                 `json:"nextLink,omitempty"`
}

// ManagementLockListResultPreparer prepares a request to retrieve the next set of results. It returns
// nil if no more results exist.
func (client ManagementLockListResult) ManagementLockListResultPreparer() (*http.Request, error) {
	if client.NextLink == nil || len(to.String(client.NextLink)) <= 0 {
		return nil, nil
	}
	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(to.String(client.NextLink)))
}

// ManagementLockObject is management lock information.
type ManagementLockObject struct {
	autorest.Response `json:"-"`
	Properties        *ManagementLockProperties `json:"properties,omitempty"`
	ID                *string                   `json:"id,omitempty"`
	Type              *string                   `json:"type,omitempty"`
	Name              *string                   `json:"name,omitempty"`
}

// ManagementLockProperties is the management lock properties.
type ManagementLockProperties struct {
	Level LockLevel `json:"level,omitempty"`
	Notes *string   `json:"notes,omitempty"`
}

// Resource is
type Resource struct {
	ID       *string             `json:"id,omitempty"`
	Name     *string             `json:"name,omitempty"`
	Type     *string             `json:"type,omitempty"`
	Location *string             `json:"location,omitempty"`
	Tags     *map[string]*string `json:"tags,omitempty"`
}

// ResourceGroupFilter is resource group filter.
type ResourceGroupFilter struct {
	TagName  *string `json:"tagName,omitempty"`
	TagValue *string `json:"tagValue,omitempty"`
}

// SubResource is
type SubResource struct {
	ID *string `json:"id,omitempty"`
}
