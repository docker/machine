/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var catalogResponses = map[string]testResponse{
	"/api/org/11111111-1111-1111-1111-111111111111":         {200, nil, orgExample},
	"/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854":     {200, nil, catalogExample},
	"/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae": {200, nil, catalogitemExample},
}

func Test_FindCatalogItem(t *testing.T) {
	cc := new(callCounter)
	ctx, err := setupTestContext(authHandler(testHandler(catalogResponses, cc)))
	if assert.NoError(t, err) {

		org, err := ctx.VDC.GetVDCOrg()
		if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {

			cat, err := org.FindCatalog("Public Catalog")
			if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {

				catitem, err := cat.FindCatalogItem("CentOS64-32bit")
				if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
					assert.Equal(t, ctx.Server.URL+"/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae", catitem.CatalogItem.HREF)
					assert.Equal(t, "id: cts-6.4-32bit", catitem.CatalogItem.Description)
				}

				_, err = cat.FindCatalogItem("INVALID")
				assert.Error(t, err)
			}
		}
	}
}

var catalogExample = `
	<?xml version="1.0" ?>
	<Catalog href="http://localhost:4444/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854" id="urn:vcloud:catalog:e8a20fdf-8a78-440c-ac71-0420db59f854" name="Public Catalog" type="application/vnd.vmware.vcloud.catalog+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
		<Link href="http://localhost:4444/api/catalog/e8a20fdf-8a78-440c-ac71-0420db59f854/metadata" rel="down" type="application/vnd.vmware.vcloud.metadata+xml"/>
		<Description>vCHS service catalog</Description>
		<CatalogItems>
			<CatalogItem href="http://localhost:4444/api/catalogItem/013d1994-f009-4c40-ac48-517fe7d952a0" id="013d1994-f009-4c40-ac48-517fe7d952a0" name="W2K12-STD-64BIT" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/05384603-e07e-4f00-a95e-776b427f22d9" id="05384603-e07e-4f00-a95e-776b427f22d9" name="W2K12-STD-R2-SQL2K14-WEB" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/1176e485-8858-4e15-94e5-ae4face605ae" id="1176e485-8858-4e15-94e5-ae4face605ae" name="CentOS64-32bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/1a729040-71b6-412c-bda9-20b9085f9882" id="1a729040-71b6-412c-bda9-20b9085f9882" name="W2K8-STD-R2-64BIT-SQL2K8-STD-R2-SP2" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/222624b5-e62a-4f5b-a2af-b33a4664005e" id="222624b5-e62a-4f5b-a2af-b33a4664005e" name="W2K12-STD-64BIT-SQL2K12-STD-SP1" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/54cb2af1-4439-48fe-85b6-4c9524930ce6" id="54cb2af1-4439-48fe-85b6-4c9524930ce6" name="Ubuntu Server 12.04 LTS (amd64 20140619)" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/693f342b-d872-41d1-983b-fd5cc2c15f7c" id="693f342b-d872-41d1-983b-fd5cc2c15f7c" name="W2K8-STD-R2-64BIT" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/8d4edd11-393f-4cda-ace4-d5b8f1548928" id="8d4edd11-393f-4cda-ace4-d5b8f1548928" name="CentOS64-64bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/bfca201c-e8f3-49f8-a828-397e16fa6cfe" id="bfca201c-e8f3-49f8-a828-397e16fa6cfe" name="W2K12-STD-R2-64BIT" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/cb508cd9-664a-4fec-8eb1-ae5934aad6ad" id="cb508cd9-664a-4fec-8eb1-ae5934aad6ad" name="W2K12-STD-64BIT-SQL2K12-WEB-SP1" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/d0be59f3-ef80-4298-bd4c-f2258a3fec37" id="d0be59f3-ef80-4298-bd4c-f2258a3fec37" name="W2K8-STD-R2-64BIT-SQL2K8-WEB-R2-SP2" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/dbbf4633-64a3-4ac1-b9e0-7f923efa3f13" id="dbbf4633-64a3-4ac1-b9e0-7f923efa3f13" name="Ubuntu Server 12.04 LTS (i386 20140619)" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/ed996ae8-3081-4e16-a7b6-4bed1c462aa4" id="ed996ae8-3081-4e16-a7b6-4bed1c462aa4" name="CentOS63-64bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/f4dc0f92-74ae-413e-8e0f-25e6568a8195" id="f4dc0f92-74ae-413e-8e0f-25e6568a8195" name="W2K12-STD-R2-SQL2K14-STD" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
			<CatalogItem href="http://localhost:4444/api/catalogItem/ff9c9b63-ca3b-4e39-ab72-7eb9049f8b05" id="ff9c9b63-ca3b-4e39-ab72-7eb9049f8b05" name="CentOS63-32bit" type="application/vnd.vmware.vcloud.catalogItem+xml"/>
		</CatalogItems>
		<IsPublished>true</IsPublished>
		<DateCreated>2013-10-15T01:14:22.370Z</DateCreated>
		<VersionNumber>60</VersionNumber>
	</Catalog>
`
