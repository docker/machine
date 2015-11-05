/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindVDCNetwork(t *testing.T) {
	cc := new(callCounter)
	responses := map[string]testResponse{
		"/api/network/44444444-4444-4444-4444-4444444444444": {200, nil, orgvdcnetExample},
	}

	ctx, err := setupTestContext(authHandler(testHandler(responses, cc)))
	if assert.NoError(t, err) {
		net, err := ctx.VDC.FindVDCNetwork("networkName")
		if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
			assert.NotNil(t, net)
			assert.Equal(t, ctx.Server.URL+"/api/network/cb0f4c9e-1a46-49d4-9fcb-d228000a6bc1", net.OrgVDCNetwork.HREF)
		}

		net, err = ctx.VDC.FindVDCNetwork("INVALID")
		assert.Error(t, err)
	}
}

func Test_GetVDCOrg(t *testing.T) {
	cc := new(callCounter)
	ctx, err := setupTestContext(authHandler(testHandler(catalogResponses, cc)))
	if assert.NoError(t, err) {
		org, err := ctx.VDC.GetVDCOrg()
		if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
			assert.Equal(t, ctx.Server.URL+"/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57", org.Org.HREF)
		}
	}
}

func Test_NewVdc(t *testing.T) {

	cc := new(callCounter)
	responses := map[string]testResponse{
		"/api/vdc/00000000-0000-0000-0000-000000000000":       {200, nil, vdcExample},
		"/api/vApp/vapp-00000000-0000-0000-0000-000000000000": {200, nil, vappExample},
	}
	ctx, err := setupTestContext(authHandler(testHandler(responses, cc)))
	if assert.NoError(t, err) {
		err := ctx.VDC.Refresh()
		if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
			vdc := ctx.VDC.Vdc
			lnk := vdc.Link[0]
			assert.Equal(t, "up", lnk.Rel)
			assert.Equal(t, "application/vnd.vmware.vcloud.org+xml", lnk.Type)
			assert.Equal(t, ctx.Server.URL+"/api/org/11111111-1111-1111-1111-111111111111", lnk.HREF)

			assert.Equal(t, "AllocationPool", vdc.AllocationModel)

			for _, v := range vdc.ComputeCapacity {
				cpu, mem := v.CPU, v.Memory
				assert.Equal(t, "MHz", cpu.Units)
				assert.Equal(t, int64(30000), cpu.Allocated)
				assert.Equal(t, int64(30000), cpu.Limit)
				assert.Equal(t, int64(15000), cpu.Reserved)
				assert.Equal(t, int64(0), cpu.Used)
				assert.Equal(t, int64(0), cpu.Overhead)

				assert.Equal(t, "MB", mem.Units)
				assert.Equal(t, int64(61440), mem.Allocated)
				assert.Equal(t, int64(61440), mem.Limit)
				assert.Equal(t, int64(61440), mem.Reserved)
				assert.Equal(t, int64(6144), mem.Used)
				assert.Equal(t, int64(95), mem.Overhead)
			}

			entity := vdc.ResourceEntities[0].ResourceEntity[0]
			assert.Equal(t, "vAppTemplate", entity.Name)
			assert.Equal(t, "application/vnd.vmware.vcloud.vAppTemplate+xml", entity.Type)
			assert.Equal(t, ctx.Server.URL+"/api/vAppTemplate/vappTemplate-22222222-2222-2222-2222-222222222222", entity.HREF)

			for _, v := range vdc.AvailableNetworks {
				for _, v2 := range v.Network {
					assert.Equal(t, "networkName", v2.Name)
					assert.Equal(t, "application/vnd.vmware.vcloud.network+xml", v2.Type)
					assert.Equal(t, ctx.Server.URL+"/api/network/44444444-4444-4444-4444-4444444444444", v2.HREF)
				}
			}

			assert.Equal(t, 0, vdc.NicQuota)
			assert.Equal(t, 20, vdc.NetworkQuota)
			assert.Equal(t, 0, vdc.UsedNetworkCount)
			assert.Equal(t, 0, vdc.VMQuota)
			assert.True(t, vdc.IsEnabled)

			for _, v := range vdc.VdcStorageProfiles {
				for _, v2 := range v.VdcStorageProfile {
					assert.Equal(t, "storageProfile", v2.Name)
					assert.Equal(t, "application/vnd.vmware.vcloud.vdcStorageProfile+xml", v2.Type)
					assert.Equal(t, ctx.Server.URL+"/api/vdcStorageProfile/88888888-8888-8888-8888-888888888888", v2.HREF)
				}
			}
		}
	}

}

func Test_FindVApp(t *testing.T) {
	cc := new(callCounter)
	responses := map[string]testResponse{
		"/api/vdc/00000000-0000-0000-0000-000000000000":       {200, nil, vdcExample},
		"/api/vApp/vapp-00000000-0000-0000-0000-000000000000": {200, nil, vappExample},
	}
	ctx, err := setupTestContext(authHandler(testHandler(responses, cc)))
	if assert.NoError(t, err) {
		_, err := ctx.VDC.FindVAppByName("myVApp")
		if assert.NoError(t, err) && assert.Equal(t, 2, cc.Pop()) {
			_, err = ctx.VDC.FindVAppByID("urn:vcloud:vapp:00000000-0000-0000-0000-000000000000")
			assert.NoError(t, err)
			assert.Equal(t, 2, cc.Pop())
		}
	}
}

var vdcExample = `
	<?xml version="1.0" ?>
	<Vdc href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000" id="urn:vcloud:vdc:00000000-0000-0000-0000-000000000000" name="M916272752-5793" status="1" type="application/vnd.vmware.vcloud.vdc+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-in stance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
	  <Link href="http://localhost:4444/api/org/11111111-1111-1111-1111-111111111111" rel="up" type="application/vnd.vmware.vcloud.org+xml"/>
	  <Link href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000/edgeGateways" rel="edgeGateways" type="application/vnd.vmware.vcloud.query.records+xml"/>
	  <AllocationModel>AllocationPool</AllocationModel>
	  <ComputeCapacity>
	    <Cpu>
	      <Units>MHz</Units>
	      <Allocated>30000</Allocated>
	      <Limit>30000</Limit>
	      <Reserved>15000</Reserved>
	      <Used>0</Used>
	      <Overhead>0</Overhead>
	    </Cpu>
	    <Memory>
	      <Units>MB</Units>
	      <Allocated>61440</Allocated>
	      <Limit>61440</Limit>
	      <Reserved>61440</Reserved>
	      <Used>6144</Used>
	      <Overhead>95</Overhead>
	    </Memory>
	  </ComputeCapacity>
	  <ResourceEntities>
	    <ResourceEntity href="http://localhost:4444/api/vAppTemplate/vappTemplate-22222222-2222-2222-2222-222222222222" name="vAppTemplate" type="application/vnd.vmware.vcloud.vAppTemplate+xml"/>
      <ResourceEntity href="http://localhost:4444/api/vApp/vapp-00000000-0000-0000-0000-000000000000" name="myVApp" type="application/vnd.vmware.vcloud.vApp+xml"/>
	  </ResourceEntities>
	  <AvailableNetworks>
	    <Network href="http://localhost:4444/api/network/44444444-4444-4444-4444-4444444444444" name="networkName" type="application/vnd.vmware.vcloud.network+xml"/>
	  </AvailableNetworks>
	  <Capabilities>
	    <SupportedHardwareVersions>
	      <SupportedHardwareVersion>vmx-10</SupportedHardwareVersion>
	    </SupportedHardwareVersions>
	  </Capabilities>
	  <NicQuota>0</NicQuota>
	  <NetworkQuota>20</NetworkQuota>
	  <UsedNetworkCount>0</UsedNetworkCount>
	  <VmQuota>0</VmQuota>
	  <IsEnabled>true</IsEnabled>
	  <VdcStorageProfiles>
	    <VdcStorageProfile href="http://localhost:4444/api/vdcStorageProfile/88888888-8888-8888-8888-888888888888" name="storageProfile" type="application/vnd.vmware.vcloud.vdcStorageProfile+xml"/>
	  </VdcStorageProfiles>
	</Vdc>
	`
