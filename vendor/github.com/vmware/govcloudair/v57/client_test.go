package v57

import (
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOAuthToken(t *testing.T) {
	authToken := "imagine this is a ridiculoulsly long kind of random string"

	tc := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Accept")
		mt, params, err := mime.ParseMediaType(ct)
		if err != nil {
			fmt.Println(err)
			rw.WriteHeader(500)
			return
		}

		if r.URL.Path == LoginPath || r.URL.Path == InstancesPath {
			if mt != "application/json" {
				rw.WriteHeader(406)
				return
			}

			if params["version"] != "5.7" {
				rw.WriteHeader(412)
				return
			}

			if r.URL.Path == LoginPath {
				un, pw, ok := r.BasicAuth()
				if !ok || un != "some user" || pw != "some password" {
					rw.WriteHeader(401)
					return
				}
				rw.Header().Set("vchs-authorization", authToken)
				rw.Header().Set("Content-Type", ct)
				rw.WriteHeader(200)
				rw.Write([]byte(`{"serviceGroupIds":["service-group-uuid-goes-here"]}`))
				return
			}

			if r.URL.Path == InstancesPath {
				rw.Header().Set("Content-Type", ct+"; class=com.vmware.vchs.sc.restapi.model.instancelisttype")
				rw.WriteHeader(200)
				rw.Write([]byte(strings.Replace(instancesJSON, "https://us-california-1-3.vchs.vmware.com", "http://"+r.Host, -1)))
				return
			}
		} else { // this should be XML and so forth
			un, pw, ok := r.BasicAuth()
			if !ok || un != "some user@org-name-uuid-goes-here" || pw != "some password" {
				rw.WriteHeader(401)
				return
			}
			rw.Header().Set("Content-Type", "application/vnd.vmware.vcloud.session+xml;version=5.11")
			rw.Header().Set("X-Vcloud-Authorization", "super-secret-cloud-auth-token")
			rw.WriteHeader(200)
			rw.Write([]byte(strings.Replace(sessionsXML, "https://us-california-1-3.vchs.vmware.com", "http://"+r.Host, -1)))
			return
		}

		rw.WriteHeader(404)

	}))
	defer tc.Close()

	os.Setenv("VCLOUDAIR_ENDPOINT", tc.URL)
	client, err := NewClient()
	if assert.NoError(t, err) {
		err = client.Authenticate("some user", "some password")

		if assert.NoError(t, err) {
			assert.Equal(t, "super-secret-cloud-auth-token", client.VCDToken)
			assert.Equal(t, authToken, client.VAToken)
		}
	}
}

var instancesJSON = `{
    "instances": [
        {
            "apiUrl": "https://storage.googleapis.com",
            "dashboardUrl": "https://us-california-1-3.vchs.vmware.com/os-g/ui/",
            "description": "Highly scalable and durable storage. Create buckets, upload and manage objects.",
            "id": "0000000000001",
            "instanceAttributes": "random-words-hex",
            "instanceVersion": "6",
            "link": [],
            "name": "Object Storage powered by Google",
            "planId": "region:us-california-1-3.vchs.vmware.com:planID:some-uuid-here-1",
            "region": "us-california-1-3.vchs.vmware.com",
            "serviceGroupId": "service-group-uuid-goes-here"
        },
        {
            "apiUrl": "https://us-california-1-3.vchs.vmware.com/api/compute/api/org/org-uuid-goes-here",
            "dashboardUrl": "https://us-california-1-3.vchs.vmware.com/api/compute/compute/ui/index.html?orgName=org-name-uuid-goes-here&serviceInstanceId=org-uuid-goes-here&servicePlan=plan-uuid-goes-here",
            "description": "Create virtual machines, and easily scale up or down as your needs change.",
            "id": "org-uuid-goes-here",
            "instanceAttributes": "{\"orgName\":\"org-name-uuid-goes-here\",\"sessionUri\":\"https://us-california-1-3.vchs.vmware.com/api/compute/api/sessions\",\"apiVersionUri\":\"https://us-california-1-3.vchs.vmware.com/api/compute/api/versions\"}",
            "instanceVersion": "1.0",
            "link": [],
            "name": "Virtual Private Cloud OnDemand",
            "planId": "region:us-california-1-3.vchs.vmware.com:planID:plan-uuid-goes-here",
            "region": "us-california-1-3.vchs.vmware.com",
            "serviceGroupId": "service-group-uuid-goes-here"
        }
    ]
}
`

var sessionsXML = `<?xml version="1.0" encoding="UTF-8"?>
<Session xmlns="http://www.vmware.com/vcloud/v1.5" org="org-name-uuid-goes-here" roles="Account Administrator" user="someone@somewhere.com" userId="urn:vcloud:user:user-uuid-goes-here" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/session" type="application/vnd.vmware.vcloud.session+xml" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.vmware.com/vcloud/v1.5 http://us-california-1-3.vchs.vmware.com/api/compute/api/v1.5/schema/master.xsd">
    <Link rel="down" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/org/" type="application/vnd.vmware.vcloud.orgList+xml"/>
    <Link rel="remove" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/session"/>
    <Link rel="down" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/admin/" type="application/vnd.vmware.admin.vcloud+xml"/>
    <Link rel="down" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/org/org-uuid-goes-here" name="org-name-uuid-goes-here" type="application/vnd.vmware.vcloud.org+xml"/>
    <Link rel="down" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/query" type="application/vnd.vmware.vcloud.query.queryList+xml"/>
    <Link rel="entityResolver" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/entity/" type="application/vnd.vmware.vcloud.entity+xml"/>
    <Link rel="down:extensibility" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/extensibility" type="application/vnd.vmware.vcloud.apiextensibility+xml"/>
    <Link rel="down" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/vchs/query?type=edgeGateway" type="application/vnd.vmware.vchs.query.records+xml"/>
    <Link rel="down" href="https://us-california-1-3.vchs.vmware.com/api/compute/api/vchs/query?type=orgVdcNetwork" type="application/vnd.vmware.vchs.query.records+xml"/>
</Session>
`
