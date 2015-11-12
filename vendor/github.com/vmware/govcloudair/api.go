/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// Package govcloudair provides a simple binding for vCloud Air REST APIs.
package govcloudair

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	types "github.com/vmware/govcloudair/types/v56"
)

// Client provides a client to vCloud Air, values can be populated automatically using the Authenticate method.
type Client interface {
	BaseURL() url.URL // HREF of the backend VDC you're using
	NewRequest(map[string]string, string, *url.URL, io.Reader) *http.Request
	DoHTTP(*http.Request) (*http.Response, error)
	Disconnect() error
}

// parseErr takes an error XML resp and returns a single string for use in error messages.
func parseErr(resp *http.Response) error {

	errBody := new(types.Error)

	// if there was an error decoding the body, just return that
	if err := decodeBody(resp, errBody); err != nil {
		return fmt.Errorf("error parsing error body for non-200 request: %s", err)
	}

	return fmt.Errorf("API Error: %d: %s", errBody.MajorErrorCode, errBody.Message)
}

// decodeBody is used to XML decode a response body
func decodeBody(resp *http.Response, out interface{}) error {

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal the XML.
	if err = xml.Unmarshal(body, &out); err != nil {
		return err
	}

	return nil
}

// checkResp wraps http.Client.Do() and verifies the request, if status code
// is 2XX it passes back the response, if it's a known invalid status code it
// parses the resultant XML error and returns a descriptive error, if the
// status code is not handled it returns a generic error with the status code.
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return resp, err
	}

	switch i := resp.StatusCode; {
	// Valid request, return the response.
	case i == 200 || i == 201 || i == 202 || i == 204:
		return resp, nil
	// Invalid request, parse the XML error returned and return it.
	case i == 400 || i == 401 || i == 403 || i == 404 || i == 405 || i == 406 || i == 409 || i == 415 || i == 500 || i == 503 || i == 504:
		return nil, parseErr(resp)
	// Unhandled response.
	default:
		return nil, fmt.Errorf("unhandled API response, please report this issue, status code: %s", resp.Status)
	}
}
