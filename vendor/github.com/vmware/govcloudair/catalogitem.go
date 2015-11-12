/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"fmt"
	"net/url"

	types "github.com/vmware/govcloudair/types/v56"
)

// CatalogItem a client catalog item
type CatalogItem struct {
	CatalogItem *types.CatalogItem
	c           Client
}

// NewCatalogItem creates a new client catalog item
func NewCatalogItem(c Client) *CatalogItem {
	return &CatalogItem{
		CatalogItem: new(types.CatalogItem),
		c:           c,
	}
}

// GetVAppTemplate gets a vApp template
func (ci *CatalogItem) GetVAppTemplate() (VAppTemplate, error) {
	url, err := url.ParseRequestURI(ci.CatalogItem.Entity.HREF)

	if err != nil {
		return VAppTemplate{}, fmt.Errorf("error decoding catalogitem response: %s", err)
	}

	req := ci.c.NewRequest(map[string]string{}, "GET", url, nil)

	resp, err := checkResp(ci.c.DoHTTP(req))
	if err != nil {
		return VAppTemplate{}, fmt.Errorf("error retreiving vapptemplate: %s", err)
	}

	cat := NewVAppTemplate(ci.c)

	if err = decodeBody(resp, cat.VAppTemplate); err != nil {
		return VAppTemplate{}, fmt.Errorf("error decoding vapptemplate response: %s", err)
	}

	// The request was successful
	return *cat, nil

}
