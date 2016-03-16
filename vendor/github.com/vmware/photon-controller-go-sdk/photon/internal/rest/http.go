// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package rest

import (
	"io"
	"net/http"
)

// Interface for abstracting away HTTP client implementation to enable testing
type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, bodyType string, body io.Reader) (resp *http.Response, err error)
	Delete(url string) (resp *http.Response, err error)
}

// Default, real implementation of HttpClient
type DefaultHttpClient struct{}

func (_ DefaultHttpClient) Get(uri string) (resp *http.Response, err error) {
	resp, err = http.Get(uri)
	return
}

func (_ DefaultHttpClient) Post(uri, bodyType string, body io.Reader) (resp *http.Response, err error) {
	resp, err = http.Post(uri, bodyType, body)
	return
}

func (_ DefaultHttpClient) Delete(uri string) (resp *http.Response, err error) {
	client := http.DefaultClient
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return
	}
	resp, err = client.Do(req)
	return
}
