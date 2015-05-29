/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwareappcatalyst

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/machine/log"
)

type VMDelete struct {
	ID string `json:"id"`
}

type VMInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type Response struct {
	Code    int    `json:"code"`    // HTTP Return code
	Message string `json:"message"` // Response message
}

type VM struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type CloneVM struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type SharedFolder struct {
	GuestPath string `json:"guestPath"`
	HostPath  string `json:"hostPath"`
	Flags     int    `json:"flags"`
}

// Client is the struct that holds information on the AppCatalyst endpoint
type Client struct {
	Endpoint url.URL     // Endpoint is the REST api endpoint AppCatalyst is listening on
	HTTP     http.Client // HttpClient is the client to use. Default will be used if not provided.
}

// NewClient returns a new client to be used with every AppCatalyst interaction
func NewClient(e string) (*Client, error) {

	u, err := url.ParseRequestURI(e)
	if err != nil {
		return &Client{}, fmt.Errorf("cannot parse endpoint, make sure it's a complete URL")
	}

	Client := Client{
		Endpoint: *u,
		HTTP:     http.Client{},
	}
	return &Client, nil
}

// GetVMSharedFolders gets shared folders information for a specific VM
func (c *Client) GetVMSharedFolders(ID string) ([]string, error) {

	s := c.Endpoint
	s.Path += "/api/vms/" + ID + "/folders"

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var folderlist []string

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &folderlist); err != nil {
		return nil, err
	}

	return folderlist, nil

}

func (c *Client) SetVMSharedFolders(ID, operation string) (*Response, error) {

	s := c.Endpoint
	s.Path += "/api/vms/" + ID + "/folders"

	b := bytes.NewBufferString(operation)

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "PATCH", s, b)

	response := new(Response)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return response, err
	}

	return response, nil

}

// CloneVM creates a clone starting from an existing VM
func (c *Client) AddVMSharedFolder(ID string, gp string, hp string, flags int) (*SharedFolder, error) {

	s := c.Endpoint
	s.Path += "/api/vms/" + ID + "/folders"

	create := SharedFolder{
		GuestPath: gp,
		HostPath:  hp,
		Flags:     flags,
	}

	out, err := json.Marshal(create)

	b := bytes.NewBufferString(string(out))

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "POST", s, b)

	sf := new(SharedFolder)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return sf, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sf, err
	}

	if err = json.Unmarshal(body, &sf); err != nil {
		return sf, err
	}

	return sf, nil

}

// GetVM gets information for a specific VM
func (c *Client) GetVMSharedFolder(ID, folderID string) (*SharedFolder, error) {

	s := c.Endpoint
	s.Path += "/api/vms/" + ID + "/folders/" + folderID

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	sf := new(SharedFolder)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return sf, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return sf, err
	}

	if err = json.Unmarshal(body, &sf); err != nil {
		return sf, err
	}

	return sf, nil

}

// GetVM gets information for a specific VM
func (c *Client) GetVM(ID string) (*VM, error) {

	s := c.Endpoint
	s.Path += "/api/vms/" + ID

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	vm := new(VM)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return vm, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return vm, err
	}

	if err = json.Unmarshal(body, &vm); err != nil {
		return vm, err
	}

	return vm, nil

}

func (c *Client) GetPowerVM(ID string) (*Response, error) {

	s := c.Endpoint
	s.Path += "/api/vms/power/" + ID

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	response := new(Response)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return response, err
	}

	return response, nil

}

func (c *Client) GetVMIPAddress(ID string) (*Response, error) {

	s := c.Endpoint
	s.Path += "/api/vms/" + ID + "/ipaddress"

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	r := new(Response)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	if err = json.Unmarshal(body, &r); err != nil {
		return r, err
	}

	return r, nil

}

func (c *Client) PowerVM(ID, operation string) (*Response, error) {

	s := c.Endpoint
	s.Path += "/api/vms/power/" + ID

	b := bytes.NewBufferString(operation)

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "PATCH", s, b)

	response := new(Response)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return response, err
	}

	return response, nil

}

func (c *Client) DeleteVM(ID string) error {

	s := c.Endpoint
	s.Path += "/api/vms/" + ID

	create := VMDelete{
		ID: ID,
	}

	out, err := json.Marshal(create)

	b := bytes.NewBufferString(string(out))

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "DELETE", s, b)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// CloneVM creates a clone starting from an existing VM
func (c *Client) CloneVM(ID, name, tag string) (*CloneVM, error) {

	s := c.Endpoint
	s.Path += "/api/vms"

	create := CloneVM{
		ID:   ID,
		Name: name,
		Tag:  tag,
	}

	out, err := json.Marshal(create)

	b := bytes.NewBufferString(string(out))

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "POST", s, b)

	clonedvm := new(CloneVM)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return clonedvm, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return clonedvm, err
	}

	if err = json.Unmarshal(body, &clonedvm); err != nil {
		return clonedvm, err
	}

	return clonedvm, nil

}

// ListVMs lists all the VMs available on AppCatalyst.
func (c *Client) ListVMs() ([]string, error) {

	s := c.Endpoint
	s.Path += "/api/vms"

	// No point in checking for errors here
	req := c.NewRequest(map[string]string{}, "GET", s, nil)

	resp, err := checkResp(c.HTTP.Do(req))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var vmlist []string

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(body, &vmlist); err != nil {
		return nil, err
	}

	return vmlist, nil

}

// NewRequest creates a new HTTP request and applies necessary auth headers if
// set.
func (c *Client) NewRequest(params map[string]string, method string, u url.URL, body io.Reader) *http.Request {

	p := url.Values{}

	// Build up our request parameters
	for k, v := range params {
		p.Add(k, v)
	}

	// Add the params to our URL
	u.RawQuery = p.Encode()

	log.Debugf("appcatalyst_driver: HTTP Debug - Method: %s URI: %s", method, u.String())
	// Build the request, no point in checking for errors here as we're just
	// passing a string version of an url.URL struct and http.NewRequest returns
	// error only if can't process an url.ParseRequestURI().
	req, _ := http.NewRequest(method, u.String(), body)

	return req

}

// parseErr takes an error resp and returns a single string for use in error
// messages.
func parseErr(resp *http.Response) error {

	errBody := new(Response)

	// if there was an error decoding the body, just return that
	if err := decodeBody(resp, errBody); err != nil {
		return fmt.Errorf("error parsing error body for non-200 request: %s", err)
	}

	return fmt.Errorf("API Response: %d: %s", errBody.Code, errBody.Message)
}

// decodeBody is used to decode a response body
func decodeBody(resp *http.Response, out interface{}) error {

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal the JSON body.
	if err = json.Unmarshal(body, &out); err != nil {
		return err
	}

	return nil
}

// checkResp wraps http.Client.Do() and verifies the request, if status code
// is 2XX it passes back the response, if it's a known invalid status code it
// parses the resultant XML error and returns a descriptive error, if the
// status code is not handled it returns a generic error with the status code.
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	// If err is already set, we can't connect to the endpoint.
	if err != nil {
		return resp, fmt.Errorf("Can't connect to AppCatalyst endpoint, make sure the REST API daemon is active")
	}

	switch i := resp.StatusCode; {
	// Valid request, return the response.
	case i == 200 || i == 201 || i == 202 || i == 204:
		return resp, nil
		// Invalid request, parse the XML error returned and return it.
	case i == 400 || i == 401 || i == 403 || i == 404 || i == 405 || i == 406 || i == 408 || i == 409 || i == 415 || i == 500 || i == 503 || i == 504:
		return nil, parseErr(resp)
	// Unhandled response.
	default:
		return nil, fmt.Errorf("unhandled API response, please report this issue, status code: %s", resp.Status)
	}
}
