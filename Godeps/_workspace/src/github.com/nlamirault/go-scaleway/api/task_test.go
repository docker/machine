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

func TestJsonTask(t *testing.T) {
	b := []byte(`{
                     "task": {
                        "description": "server_poweroff",
                        "href_from": "/servers/741db378-6b87-46d4-a8c5-4e46a09ab1f8/action",
                        "id": "a8a1775c-0dda-4f52-87b2-4e8101d68d6e",
                        "progress": 0,
                        "status": "pending"
                      }
                   }`)
	var response TaskResponse
	err := json.Unmarshal(b, &response)
	if err != nil {
		t.Fatal(err)
	}
	description := "server_poweroff"
	if response.Task.Description != description {
		t.Fatalf("Expected %s, got %s",
			description, response.Task.Description)
	}
	id := "a8a1775c-0dda-4f52-87b2-4e8101d68d6e"
	if response.Task.ID != id {
		t.Fatalf("Expected %s, got %s",
			id, response.Task.ID)
	}

}
