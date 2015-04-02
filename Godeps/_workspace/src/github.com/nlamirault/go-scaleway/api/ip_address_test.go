// Copyright (C) 2015  Nicolas Lamirault <nicolas.lamirault@gmail.com>

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

func TestJsonIPAddress(t *testing.T) {
	b := []byte(`{
  "ip": {
    "address": "212.47.226.88",
    "id": "b50cd740-892d-47d3-8cbf-88510ef626e7",
    "organization": "000a115d-2852-4b0a-9ce8-47f1134ba95a",
    "server": null
  }
}`)
	var response IPAddressResponse
	err := json.Unmarshal(b, &response)
	if err != nil {
		t.Fatal(err)
	}
	if response.IPAddress.Address != "212.47.226.88" {
		t.Fatalf("Invalid IP Address %s",
			response.IPAddress.Address)
	}
}
