// Copyright (C) 2015 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ApiError struct {
	StatusCode int
	Header     http.Header
	Body       string
	URL        *url.URL
}

func newApiError(resp *http.Response) *ApiError {
	p, _ := ioutil.ReadAll(resp.Body)
	return &ApiError{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       string(p),
		URL:        resp.Request.URL,
	}
}

// ApiError supports the error interface
func (e ApiError) Error() string {
	return fmt.Sprintf("Get %s returned status %d, %s",
		e.URL, e.StatusCode, e.Body)
}
