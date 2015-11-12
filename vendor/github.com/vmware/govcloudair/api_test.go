/*
 * @Author: frapposelli, casualjim
 * @Project: govcloudair
 * @Filename: api_test.go
 * @Last Modified by: casualjim
 */

package govcloudair

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/vmware/govcloudair/v56"
)

var authheader = map[string]string{"x-vchs-authorization": "012345678901234567890123456789"}

type testContext struct {
	Server *httptest.Server
	Client Client
	VDC    *Vdc
	VApp   *VApp
}

type testResponse struct {
	Code    int
	Headers map[string]string
	Body    string
}

type callCounter struct {
	current int32
}

func (cc *callCounter) Inc() {
	cc.current++
}

func (cc *callCounter) Pop() int {
	defer func() { cc.current = 0 }()
	return int(cc.current)
}

func testHandler(responses map[string]testResponse, callCount *callCounter) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if resp, ok := responses[r.URL.Path]; ok {
			callCount.Inc()
			for k, v := range resp.Headers {
				rw.Header().Add(k, v)
			}
			rw.WriteHeader(resp.Code)
			rw.Write([]byte(strings.Replace(resp.Body, "localhost:4444", r.Host, -1)))
			return
		}

		rw.WriteHeader(http.StatusNotFound)
	})
}

var authRequests = map[string]testResponse{
	"/api/vchs/sessions":                                                                                            {201, authheader, vaauthorization},
	"/api/vchs/services":                                                                                            {200, nil, vaservices},
	"/api/vchs/compute/00000000-0000-0000-0000-000000000000":                                                        {200, nil, vacompute},
	"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": {201, nil, vabackend},
	"/api/vdc/00000000-0000-0000-0000-000000000000":                                                                 {200, nil, vdcExample},
}

func authHandler(handler http.Handler) http.Handler {
	cnt := 0
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// the login story is 5 requests, after that we're definitely doing actual test requests
		if resp, ok := authRequests[r.URL.Path]; ok && cnt < 5 {
			cnt++
			for k, v := range resp.Headers {
				rw.Header().Add(k, v)
			}
			rw.WriteHeader(resp.Code)
			resp := strings.Replace(resp.Body, "localhost:4444", r.Host, -1) + "\n"
			rw.Write([]byte(resp))
			return
		}

		handler.ServeHTTP(rw, r)
	})
}

func setupTestContext(handler http.Handler) (testContext, error) {
	serv := httptest.NewServer(handler)

	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err := v56.NewClient()
	if err != nil {
		return testContext{}, err
	}

	err = client.Authenticate("username", "password", "CI123456-789", "VDC12345-6789")
	if err != nil {
		return testContext{}, err
	}

	vdc, err := RetrieveVDC(client)
	if err != nil {
		return testContext{}, err
	}

	vapp := NewVApp(client)

	return testContext{
		Server: serv,
		Client: client,
		VDC:    vdc,
		VApp:   vapp,
	}, nil
}

// status: 404
var notfoundErr = `
	<html>
		<head><title>404 Not Found</title></head>
		<body bgcolor="white">
			<center><h1>404 Not Found</h1></center>
			<hr><center>nginx/1.4.6 (Ubuntu)</center>
		</body>
	</html>
	`

// status: 201
var vaauthorization = `
	<?xml version="1.0" ?>
	<Session href="http://localhost:4444/api/vchs/session" type="application/xml;class=vnd.vmware.vchs.session" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	    <Link href="http://localhost:4444/api/vchs/services" rel="down" type="application/xml;class=vnd.vmware.vchs.servicelist"/>
	    <Link href="http://localhost:4444/api/vchs/session" rel="remove"/>
	</Session>
	`
var vaauthorizationErr = `
	<?xml version="1.0" ?>
	<Session href="http://localhost:4444/api/vchs/session" type="application/xml;class=vnd.vmware.vchs.session" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	    <Link href="http://localhost:4444/api/vchs/session" rel="remove"/>
	</Session>
	`

// status: 200
var vaservices = `
	<?xml version="1.0" ?>
	<Services href="http://localhost:4444/api/vchs/services" type="application/xml;class=vnd.vmware.vchs.servicelist" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	    <Service href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000" region="US - Anywhere" serviceId="CI123456-789" serviceType="compute:vpc" type="application/xml;class=vnd.vmware.vchs.compute"/>
	</Services>
	`

// status: 200
var vacompute = `
	<?xml version="1.0" ?>
	<Compute href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000" serviceId="CI123456-789" serviceType="compute:vpc" type="application/xml;class=vnd.vmware.vchs.compute" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	    <Link href="http://localhost:4444/api/vchs/services" name="services" rel="up" type="application/xml;class=vnd.vmware.vchs.servicelist"/>
	    <VdcRef href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000" name="VDC12345-6789" status="Active" type="application/xml;class=vnd.vmware.vchs.vdcref">
	        <Link href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" name="VDC12345-6789" rel="down" type="application/xml;class=vnd.vmware.vchs.vcloudsession"/>
	    </VdcRef>
	</Compute>
	`

// status: 201
var vabackend = `
<?xml version="1.0" ?>
<VCloudSession href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" name="CI123456-789-session" type="application/xml;class=vnd.vmware.vchs.vcloudsession" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <Link href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000" name="vdc" rel="up" type="application/xml;class=vnd.vmware.vchs.vdcref"/>
    <VdcLink authorizationHeader="x-vcloud-authorization" authorizationToken="01234567890123456789012345678901" href="http://localhost:4444/api/vdc/00000000-0000-0000-0000-000000000000" name="CI123456-789" rel="vcloudvdc" type="application/vnd.vmware.vcloud.vdc+xml"/>
</VCloudSession>
	`

// status: 201
var vabackendErr = `
<?xml version="1.0" ?>
<VCloudSession href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession" name="CI123456-789-session" type="application/xml;class=vnd.vmware.vchs.vcloudsession" xmlns="http://www.vmware.com/vchs/v5.6" xmlns:tns="http://www.vmware.com/vchs/v5.6" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <Link href="http://localhost:4444/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000" name="vdc" rel="up" type="application/xml;class=vnd.vmware.vchs.vdcref"/>
    <VdcLink authorizationHeader="x-vcloud-authorization" authorizationToken="01234567890123456789012345678901" href="http://$£$%£%$:4444/api/vdc/00000000-0000-0000-0000-000000000000" name="CI123456-789" rel="vcloudvdc" type="application/vnd.vmware.vcloud.vdc+xml"/>
</VCloudSession>
	`

var vcdError = `
<Error xmlns="http://www.vmware.com/vcloud/v1.5" message="Error Message" majorErrorCode="500" minorErrorCode="Server Error" vendorSpecificErrorCode="NoSpecificError" stackTrace="Hello my name is Stack Trace"/>
	`
