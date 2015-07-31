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
	"encoding/json"
	//"fmt"
	"testing"
)

func TestJSONUser(t *testing.T) {
	b := []byte(`{
                     "user": {"phone_number": "+336666666", "firstname": "Foo", "lastname": "Bar", "creation_date": "2014-12-30T16:26:46.679327+00:00", "ssh_public_keys": [{"key": "ssh-rsa AAAAB3NzaC1y5eIGsLPxcYW foo@foobar"}], "id": "aaa31895-520a-4ab7-9707-8bc1819a9e19", "organizations": [{"id": "4a3b-4ccc-88f3-b65e3f31fb75", "name": "foo.bar@gmail.com"}], "modification_date": "2015-02-09T10:29:38.789162+00:00", "roles": [{"organization": {"id": "4a3b-4ccc-88f3-b65e3f31fb75", "name": "foo.bar@gmail.com"}, "role": "manager"}], "fullname": "Foo Bar", "email": "foo.bar@gmail.com"}
                   }`)
	var response UserResponse
	err := json.Unmarshal(b, &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.User.PhoneNumber != "+336666666" {
		t.Errorf("Invalid user phonenumber")
	}
	if response.User.Firstname != "Foo" {
		t.Errorf("Invalid user firstname")
	}
	if response.User.Lastname != "Bar" {
		t.Errorf("Invalid user lastname")
	}

}
