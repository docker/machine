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
	"net/http"
	"net/url"
)

// Error represents an HTTP error
type Error struct {
	StatusCode int
	Header     http.Header
	Message    string
	URL        *url.URL
}

func newAPIError(resp *http.Response, body string) *Error {
	return &Error{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Message:    body,
		URL:        resp.Request.URL,
	}
}

// Error supports the error interface
func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}
