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
// "encoding/json"
)

// Task represents an asynchronous task
type Task struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	HrefFrom    string `json:"href_from,omitempty"`
	Progress    int    `json:"progress,omitempty"`
	Status      string `json:"status,omitempty"`
}

// TaskResponse represents a JSON response of task
type TaskResponse struct {
	Task Task `json:"task,omitempty"`
}

// // GetTaskFromJSON load bytes and return a TaskResponse
// func GetTaskFromJSON(b []byte) (*TaskResponse, error) {
// 	var response TaskResponse
// 	err := json.Unmarshal(b, &response)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &response, nil
// }
