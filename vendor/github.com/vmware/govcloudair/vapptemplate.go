/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	types "github.com/vmware/govcloudair/types/v56"
)

// VAppTemplate client
type VAppTemplate struct {
	VAppTemplate *types.VAppTemplate
	c            Client
}

// NewVAppTemplate create a new VAppTemplate client
func NewVAppTemplate(c Client) *VAppTemplate {
	return &VAppTemplate{
		VAppTemplate: new(types.VAppTemplate),
		c:            c,
	}
}
