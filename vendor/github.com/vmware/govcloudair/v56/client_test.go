/*
 * @Author: frapposelli, casualjim
 * @Project: govcloudair
 * @Filename: api_test.go
 * @Last Modified by: casualjim
 */

package v56

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var authheader = map[string]string{"x-vchs-authorization": "012345678901234567890123456789"}

type testContext struct {
	Server *httptest.Server
	Client *Client
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

	client, err := NewClient()
	if err != nil {
		return testContext{}, err
	}

	err = client.Authenticate("username", "password", "CI123456-789", "VDC12345-6789")
	if err != nil {
		return testContext{}, err
	}

	// vapp := NewVApp(client)

	return testContext{
		Server: serv,
		Client: client,
		// VDC:    vdc,
		// VApp:   vapp,
	}, nil
}

func TestClient_vaauthorize(t *testing.T) {
	cc := new(callCounter)
	serv := httptest.NewServer(testHandler(map[string]testResponse{
		"/api/vchs/sessions": {201, authheader, vaauthorization},
	}, cc))
	// Set up a working client
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err := NewClient()
	if !assert.NoError(t, err) {
		return
	}

	// Set up a correct conversation
	_, err = client.vaauthorize("username", "password")
	assert.Equal(t, 1, cc.Pop())

	// Test if token is correctly set on client.
	assert.Equal(t, "012345678901234567890123456789", client.VAToken)
	serv.Close()

	// Test client errors
	testError := func(resp testResponse) bool {
		serv = httptest.NewServer(testHandler(map[string]testResponse{
			"/api/vchs/sessions": resp,
		}, cc))
		os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")
		client, err := NewClient()
		if !assert.NoError(t, err) {
			return false
		}
		_, err = client.vaauthorize("username", "password")
		return assert.Error(t, err)
	}

	// Test a correct response with a wrong status code
	if !testError(testResponse{404, authheader, notfoundErr}) {
		return
	}
	// Test an API error
	if !testError(testResponse{500, authheader, vcdError}) {
		return
	}
	// Test an API response that doesn't contain the param we're looking for.
	if !testError(testResponse{200, authheader, vaauthorizationErr}) {
		return
	}
	// Test an un-parsable response.
	if !testError(testResponse{200, authheader, notfoundErr}) {
		return
	}

}

func TestClient_vaacquireservice(t *testing.T) {
	cc := new(callCounter)
	serv := httptest.NewServer(testHandler(map[string]testResponse{
		"/api/vchs/services": {200, nil, vaservices},
	}, cc))
	// Set up a working client
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err := NewClient()
	if !assert.NoError(t, err) {
		return
	}
	client.VAToken = "012345678901234567890123456789"

	// Test a correct conversation
	aus, _ := url.ParseRequestURI(serv.URL + "/api/vchs/services")
	vacomputehref, err := client.vaacquireservice(aus, "CI123456-789")
	if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
		assert.Equal(t, serv.URL+"/api/vchs/compute/00000000-0000-0000-0000-000000000000", vacomputehref.String())
		assert.Equal(t, "US - Anywhere", client.Region)
	}

	serv.Close()

	// Test client errors
	testError := func(param string, resp testResponse) bool {
		serv = httptest.NewServer(testHandler(map[string]testResponse{
			"/api/vchs/services": resp,
		}, cc))
		os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")
		client, err := NewClient()
		if !assert.NoError(t, err) {
			return false
		}
		client.VAToken = "012345678901234567890123456789"
		aus, _ := url.ParseRequestURI(serv.URL + "/api/vchs/services")
		_, err = client.vaacquireservice(aus, param)
		return assert.Error(t, err)
	}

	// Test a 404
	if !testError("CI123456-789", testResponse{404, nil, notfoundErr}) {
		return
	}

	// Test an API error
	if !testError("CI123456-789", testResponse{500, nil, vcdError}) {
		return
	}

	// Test an unknown Compute ID
	if !testError("NOTVALID-789", testResponse{200, nil, vaservices}) {
		return
	}

	// Test an un-parsable response
	if !testError("CI123456-789", testResponse{200, nil, notfoundErr}) {
		return
	}

}

func TestClient_vaacquirecompute(t *testing.T) {
	cc := new(callCounter)
	serv := httptest.NewServer(testHandler(map[string]testResponse{
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000": {200, nil, vacompute},
	}, cc))
	// Set up a working client
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err := NewClient()
	if !assert.NoError(t, err) {
		return
	}
	client.VAToken = "012345678901234567890123456789"
	client.Region = "US - Anywhere"

	auc, _ := url.ParseRequestURI(serv.URL + "/api/vchs/compute/00000000-0000-0000-0000-000000000000")
	vavdchref, err := client.vaacquirecompute(auc, "VDC12345-6789")
	if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
		assert.Equal(t, serv.URL+"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession", vavdchref.String())
	}

	// Test client errors
	testError := func(param string, resp testResponse) bool {
		serv = httptest.NewServer(testHandler(map[string]testResponse{
			"/api/vchs/compute/00000000-0000-0000-0000-000000000000": resp,
		}, cc))
		os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")
		client, err := NewClient()
		if !assert.NoError(t, err) {
			return false
		}
		client.VAToken = "012345678901234567890123456789"
		client.Region = "US - Anywhere"
		auc, _ := url.ParseRequestURI(serv.URL + "/api/vchs/compute/00000000-0000-0000-0000-000000000000")
		_, err = client.vaacquirecompute(auc, param)
		return assert.Error(t, err)
	}

	// Test a 404
	if !testError("VDC12345-6789", testResponse{404, nil, notfoundErr}) {
		return
	}

	// Test an API error
	if !testError("VDC12345-6789", testResponse{500, nil, vcdError}) {
		return
	}

	// Test an unknown VDC ID
	if !testError("INVALID-6789", testResponse{200, nil, vacompute}) {
		return
	}

	// Test an un-parsable response
	if !testError("VDC12345-6789", testResponse{200, nil, notfoundErr}) {
		return
	}

}

func TestClient_vagetbackendauth(t *testing.T) {
	cc := new(callCounter)
	serv := httptest.NewServer(testHandler(map[string]testResponse{
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": {201, nil, vabackend},
	}, cc))
	// Set up a working client
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err := NewClient()
	if !assert.NoError(t, err) {
		return
	}
	client.VAToken = "012345678901234567890123456789"
	client.Region = "US - Anywhere"

	aucs, _ := url.ParseRequestURI(serv.URL + "/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession")
	err = client.vagetbackendauth(aucs, "CI123456-789")
	if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
		assert.Equal(t, "01234567890123456789012345678901", client.VCDToken)
		assert.Equal(t, "x-vcloud-authorization", client.VCDAuthHeader)
		bu := client.BaseURL()
		assert.Equal(t, serv.URL+"/api/vdc/00000000-0000-0000-0000-000000000000", (&bu).String())
	}

	// Test client errors
	testError := func(param string, resp testResponse) bool {
		serv = httptest.NewServer(testHandler(map[string]testResponse{
			"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": resp,
		}, cc))
		os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")
		client, err := NewClient()
		if !assert.NoError(t, err) {
			return false
		}
		client.VAToken = "012345678901234567890123456789"
		client.Region = "US - Anywhere"
		aucs, _ := url.ParseRequestURI(serv.URL + "/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession")
		err = client.vagetbackendauth(aucs, param)
		return assert.Error(t, err)
	}

	// Test a 404
	if !testError("VDC12345-6789", testResponse{404, nil, notfoundErr}) {
		return
	}

	// Test an API error
	if !testError("VDC12345-6789", testResponse{500, nil, vcdError}) {
		return
	}

	// Test an unknown backend VDC IC
	if !testError("INVALID-6789", testResponse{200, nil, vabackend}) {
		return
	}

	// Test an un-parsable response
	if !testError("VDC12345-6789", testResponse{201, nil, notfoundErr}) {
		return
	}
	// Test a botched backend VDC IC
	if !testError("VDC12345-6789", testResponse{201, nil, vabackendErr}) {
		return
	}

}

//// Env variable tests

func TestClient_vaauthorize_env(t *testing.T) {
	cc := new(callCounter)
	os.Setenv("VCLOUDAIR_USERNAME", "username")
	os.Setenv("VCLOUDAIR_PASSWORD", "password")

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cc.Inc()
		rw.Header().Add("x-vchs-authorization", "012345678901234567890123456789")
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(strings.Replace(vaauthorization, "localhost:4444", r.Host, -1)))
	}))

	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err := NewClient()
	if assert.NoError(t, err) {
		_, err = client.vaauthorize("", "")
		if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
			assert.Equal(t, "012345678901234567890123456789", client.VAToken)
		}
	}

}

func TestClient_vaacquireservice_env(t *testing.T) {
	cc := new(callCounter)
	os.Setenv("VCLOUDAIR_COMPUTEID", "CI123456-789")

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cc.Inc()
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(strings.Replace(vaservices, "localhost:4444", r.Host, -1)))
	}))

	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")
	client, err := NewClient()
	if !assert.NoError(t, err) {
		return
	}
	client.VAToken = "012345678901234567890123456789"

	aus, _ := url.ParseRequestURI(serv.URL + "/api/vchs/services")
	vacomputehref, err := client.vaacquireservice(aus, "")
	if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
		assert.Equal(t, serv.URL+"/api/vchs/compute/00000000-0000-0000-0000-000000000000", vacomputehref.String())
		assert.Equal(t, "US - Anywhere", client.Region)
	}
}

func TestClient_vaacquirecompute_env(t *testing.T) {
	cc := new(callCounter)
	os.Setenv("VCLOUDAIR_VDCID", "VDC12345-6789")

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cc.Inc()
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(strings.Replace(vacompute, "localhost:4444", r.Host, -1)))
	}))

	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")
	client, err := NewClient()
	if !assert.NoError(t, err) {
		return
	}
	client.VAToken = "012345678901234567890123456789"
	client.Region = "US - Anywhere"

	auc, _ := url.ParseRequestURI(serv.URL + "/api/vchs/compute/00000000-0000-0000-0000-000000000000")
	vavdchref, err := client.vaacquirecompute(auc, "")
	if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
		assert.Equal(t, serv.URL+"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession", vavdchref.String())
	}
}

func TestClient_vagetbackendauth_env(t *testing.T) {
	cc := new(callCounter)
	os.Setenv("VCLOUDAIR_COMPUTEID", "CI123456-789")

	serv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		cc.Inc()
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(strings.Replace(vabackend, "localhost:4444", r.Host, -1)))
	}))

	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")
	client, err := NewClient()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	client.VAToken = "012345678901234567890123456789"
	client.Region = "US - Anywhere"

	aucs, _ := url.ParseRequestURI(serv.URL + "/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession")
	err = client.vagetbackendauth(aucs, "")
	if assert.NoError(t, err) && assert.Equal(t, 1, cc.Pop()) {
		assert.Equal(t, "01234567890123456789012345678901", client.VCDToken)
		assert.Equal(t, "x-vcloud-authorization", client.VCDAuthHeader)
		bu := client.BaseURL()
		assert.Equal(t, serv.URL+"/api/vdc/00000000-0000-0000-0000-000000000000", bu.String())
	}
}

func TestClient_NewClient(t *testing.T) {

	var err error
	os.Setenv("VCLOUDAIR_ENDPOINT", "")
	if _, err = NewClient(); err != nil {
		t.Fatalf("err: %v", err)
	}

	os.Setenv("VCLOUDAIR_ENDPOINT", "💩")
	if _, err = NewClient(); err == nil {
		t.Fatalf("err: %v", err)
	}

}

func TestClient_Disconnect(t *testing.T) {
	cc := new(callCounter)
	ctx, ok := makeClient(t, testHandler(map[string]testResponse{"/api/vchs/session": {201, nil, ""}}, cc))
	if assert.True(t, ok) {
		err := ctx.Client.Disconnect()
		assert.NoError(t, err)
		assert.Equal(t, 1, cc.Pop())
	}
}

func TestClient_Authenticate(t *testing.T) {

	cc := new(callCounter)
	responses := map[string]testResponse{
		"/api/vchs/sessions":                                                                                            {401, nil, vcdError},
		"/api/vchs/services":                                                                                            {200, nil, vaservices},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000":                                                        {200, nil, vacompute},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": {201, nil, vabackend},
		// "/api/vdc/00000000-0000-0000-0000-000000000000":                                                                 {200, nil, vdcExample},
	}
	serv := httptest.NewServer(testHandler(responses, cc))
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err := NewClient()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Botched auth
	err = client.Authenticate("username", "password", "CI123456-789", "VDC12345-6789")
	assert.Error(t, err)
	assert.Equal(t, 1, cc.Pop())

	// Botched services
	responses = map[string]testResponse{
		"/api/vchs/sessions":                                                                                            {201, authheader, vaauthorization},
		"/api/vchs/services":                                                                                            {500, nil, vcdError},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000":                                                        {200, nil, vacompute},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": {201, nil, vabackend},
	}
	serv = httptest.NewServer(testHandler(responses, cc))
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err = NewClient()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	err = client.Authenticate("username", "password", "CI123456-789", "VDC12345-6789")
	assert.Error(t, err)
	assert.Equal(t, 2, cc.Pop())

	// Botched compute
	responses = map[string]testResponse{
		"/api/vchs/sessions":                                                                                            {201, authheader, vaauthorization},
		"/api/vchs/services":                                                                                            {200, nil, vaservices},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000":                                                        {500, nil, vcdError},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": {201, nil, vabackend},
	}

	serv = httptest.NewServer(testHandler(responses, cc))
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err = NewClient()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	err = client.Authenticate("username", "password", "CI123456-789", "VDC12345-6789")
	assert.Error(t, err)
	assert.Equal(t, 3, cc.Pop())

	// Botched backend
	responses = map[string]testResponse{
		"/api/vchs/sessions":                                                                                            {201, authheader, vaauthorization},
		"/api/vchs/services":                                                                                            {200, nil, vaservices},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000":                                                        {200, nil, vacompute},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": {500, nil, vcdError},
	}
	serv = httptest.NewServer(testHandler(responses, cc))
	os.Setenv("VCLOUDAIR_ENDPOINT", serv.URL+"/api")

	client, err = NewClient()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	err = client.Authenticate("username", "password", "CI123456-789", "VDC12345-6789")
	assert.Error(t, err)
	assert.Equal(t, 13, cc.Pop())

}

func makeClient(t testing.TB, handler http.Handler) (testContext, bool) {
	ctx, err := setupTestContext(authHandler(handler))
	if !assert.NoError(t, err) {
		return testContext{}, false
	}
	if !assert.Equal(t, "012345678901234567890123456789", ctx.Client.VAToken) {
		return testContext{}, false
	}
	if !assert.Equal(t, "US - Anywhere", ctx.Client.Region) {
		return testContext{}, false
	}
	if !assert.Equal(t, "01234567890123456789012345678901", ctx.Client.VCDToken) {
		return testContext{}, false
	}
	if !assert.Equal(t, "x-vcloud-authorization", ctx.Client.VCDAuthHeader) {
		return testContext{}, false
	}
	bu := ctx.Client.BaseURL()
	if !assert.Equal(t, ctx.Server.URL+"/api/vdc/00000000-0000-0000-0000-000000000000", bu.String()) {
		return testContext{}, false
	}

	return ctx, true
}

func TestClient_parseErr(t *testing.T) {
	// I'M A TEAPOT!
	responses := map[string]testResponse{
		"/api/vchs/sessions":                                                                                            {201, authheader, vaauthorization},
		"/api/vchs/services":                                                                                            {200, nil, vaservices},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000":                                                        {200, nil, vacompute},
		"/api/vchs/compute/00000000-0000-0000-0000-000000000000/vdc/00000000-0000-0000-0000-000000000000/vcloudsession": {418, nil, notfoundErr},
	}
	cc := new(callCounter)
	_, err := setupTestContext(testHandler(responses, cc))
	assert.Error(t, err)
	assert.Equal(t, 13, cc.Pop())
}

func TestClient_NewRequest(t *testing.T) {
	c, _ := makeClient(t, http.NotFoundHandler())

	params := map[string]string{
		"foo": "bar",
		"baz": "bar",
	}

	uri, _ := url.ParseRequestURI(c.Server.URL + "/api/bar")
	req := c.Client.NewRequest(params, "POST", uri, nil)
	encoded := req.URL.Query()

	assert.Equal(t, "bar", encoded.Get("foo"))
	assert.Equal(t, "bar", encoded.Get("baz"))
	assert.Equal(t, c.Server.URL+"/api/bar?baz=bar&foo=bar", req.URL.String())
	assert.Equal(t, "01234567890123456789012345678901", req.Header.Get("x-vcloud-authorization"))
	assert.Equal(t, "POST", req.Method)
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
