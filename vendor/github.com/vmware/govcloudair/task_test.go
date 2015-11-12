/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcloudair

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WaitTaskCompletion(t *testing.T) {
	cc := new(callCounter)
	responses := map[string]testResponse{
		"/api/vApp/vapp-00000000-0000-0000-0000-000000000000/action/deploy": {200, nil, taskExample},
		"/api/task/1b8f926c-eff5-4bea-9b13-4e49bdd50c05":                    {200, nil, taskExample},
	}

	ctx, err := setupTestContext(authHandler(testHandler(responses, cc)))
	if assert.NoError(t, err) {
		xmlTxt := strings.Replace(vappExample, "http://localhost:4444", ctx.Server.URL, -1)
		if assert.NoError(t, xml.Unmarshal([]byte(xmlTxt), ctx.VApp.VApp)) {
			task, err := ctx.VApp.Deploy()
			if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
				err := task.WaitTaskCompletion()
				assert.NoError(t, err)
				assert.Equal(t, 1, cc.Pop())
			}
		}
	}
}

var taskExample = `
<Task cancelRequested="false" endTime="2014-11-10T09:09:31.483Z" expiryTime="2015-02-08T09:09:16.627Z" href="http://localhost:4444/api/task/1b8f926c-eff5-4bea-9b13-4e49bdd50c05" id="urn:vcloud:task:1b8f926c-eff5-4bea-9b13-4e49bdd50c05" name="task" operation="Composed Virtual Application Test API GO4(fdb86157-2e1f-4889-9942-0463836d10e1)" operationName="vdcComposeVapp" serviceNamespace="com.vmware.vcloud" startTime="2014-11-10T09:09:16.627Z" status="success" type="application/vnd.vmware.vcloud.task+xml" xmlns="http://www.vmware.com/vcloud/v1.5" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://10.6.32.3/api/v1.5/schema/master.xsd">
  <Owner href="http://localhost:4444/api/vApp/vapp-fdb86157-2e1f-4889-9942-0463836d10e1" name="Test API GO4" type="application/vnd.vmware.vcloud.vApp+xml"/>
  <User href="http://localhost:4444/api/admin/user/d8ac278a-5b49-4c85-9a81-468838e89eb9" name="frapposelli1@gts-vchs.com" type="application/vnd.vmware.admin.user+xml"/>
  <Organization href="http://localhost:4444/api/org/23bd2339-c55f-403c-baf3-13109e8c8d57" name="M916272752-5793" type="application/vnd.vmware.vcloud.org+xml"/>
  <Progress>100</Progress>
  <Details/>
</Task>
	`
