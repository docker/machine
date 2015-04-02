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

	log "github.com/Sirupsen/logrus"
)

// Volume represents a disk
type Volume struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Type         string `json:"volume_type,omitempty"`
	Size         int64  `json:"size,omitempty"`
	Server       Server `json:"server,omitempty"`
}

// VolumeResponse represents JSON response of volume
type VolumeResponse struct {
	Volume Volume `json:"volume,omitempty"`
}

// VolumesResponse represents a list of volumes in JSON
type VolumesResponse struct {
	Volumes []Volume
}

// GetVolumeFromJSON load bytes//  and return a VolumesResponse
// func GetVolumeFromJSON(b []byte) (*VolumeResponse, error) {
// 	var response VolumeResponse
// 	err := json.Unmarshal(b, &response)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &response, nil
// }

// // GetVolumesFromJSON load bytes and return a VolumesResponse
// func GetVolumesFromJSON(b []byte) (*VolumesResponse, error) {
// 	var response VolumesResponse
// 	err := json.Unmarshal(b, &response)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &response, nil
// }<

func (v Volume) Display() {
	log.Infof("Id     : %s", v.ID)
	log.Infof("Name   : %s", v.Name)
	log.Infof("Type   : %s", v.Type)
	log.Infof("Size   : %d", v.Size)
	log.Infof("Server : %s %s", v.Server.ID, v.Server.Name)
}
