/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var catalogItemsResponses = map[string]testResponse{
	"/api/org/11111111-1111-1111-1111-111111111111":                       {200, nil, orgExample},
	"/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854":                   {200, nil, catalogExample},
	"/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae":               {200, nil, catalogitemExample},
	"/api/vAppTemplate/vappTemplate-40cb9721-5f1a-44f9-b5c3-98c5f518c4f5": {200, nil, vapptemplateExample},
}

func Test_GetVAppTemplate(t *testing.T) {
	cc := new(callCounter)
	ctx, err := setupTestContext(authHandler(testHandler(catalogItemsResponses, cc)))
	if assert.NoError(t, err) {

		// Get the Org populated
		org, err := ctx.VDC.GetVDCOrg()
		if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {

			// Populate Catalog
			cat, err := org.FindCatalog("Public Catalog")
			if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {

				// Populate Catalog Item
				catitem, err := cat.FindCatalogItem("CentOS64-32bit")
				if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {

					// Get VAppTemplate
					vapptemplate, err := catitem.GetVAppTemplate()
					if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
						assert.Equal(t, ctx.Server.URL+"/api/vAppTemplate/vappTemplate-40cb9721-5f1a-44f9-b5c3-98c5f518c4f5", vapptemplate.VAppTemplate.HREF)
						assert.Equal(t, "CentOS64-32bit", vapptemplate.VAppTemplate.Name)
						assert.Equal(t, "id: cts-6.4-32bit", vapptemplate.VAppTemplate.Description)
					}
				}
			}
		}
	}
}

var catalogitemExample = `
	<?xml version="1.0" ?>
	<CatalogItem href="http://localhost:4444/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae" id="urn:vcloud:catalogitem:1176e485-8858-4e15-94e5-ae4face605ae" name="CentOS64-32bit" size="0" type="application/vnd.vmware.vcloud.catalogItem+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
		<Link href="http://localhost:4444/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854" rel="up" type="application/vnd.vmware.vcloud.catalog+xml"/>
		<Link href="http://localhost:4444/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae/metadata" rel="down" type="application/vnd.vmware.vcloud.metadata+xml"/>
		<Description>id: cts-6.4-32bit</Description>
		<Entity href="http://localhost:4444/api/vAppTemplate/vappTemplate-40cb9721-5f1a-44f9-b5c3-98c5f518c4f5" name="CentOS64-32bit" type="application/vnd.vmware.vcloud.vAppTemplate+xml"/>
		<DateCreated>2014-06-04T21:06:43.750Z</DateCreated>
		<VersionNumber>4</VersionNumber>
	</CatalogItem>
`
