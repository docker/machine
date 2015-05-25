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

	//log "github.com/Sirupsen/logrus"
	"github.com/nlamirault/go-scaleway/log"
)

// IPAddress represents an IP entity
type IPAddress struct {
	ID           string `json:"id,omitempty"`
	Address      string `json:"address,omitempty"`
	Organization string `json:"organization,omitempty"`
	Server       string `json:"server,omitempty"`
}

// IPAddressResponse represents JSON response of IP address
type IPAddressResponse struct {
	IPAddress `json:"ip,omitempty"`
}

// IPAddressesResponse represents JSON response of list of IP address
type IPAddressesResponse struct {
	IPAddresses []IPAddress
}

// GetIPAddressFromJSON load bytes and return a IPAddressResponse
func GetIPAddressFromJSON(b []byte) (*IPAddressResponse, error) {
	var response IPAddressResponse
	err := json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetIPAddressesFromJSON load bytes and return a IPAddressesResponse
func GetIPAddressesFromJSON(b []byte) (*IPAddressesResponse, error) {
	var response IPAddressesResponse
	err := json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Display log the IPAddress caracteristics
func (ip IPAddress) Display() {
	log.Infof("Id           : %s", ip.ID)
	log.Infof("Address      : %s", ip.Address)
	log.Infof("Organization : %s", ip.Organization)
	log.Infof("Server       : %s", ip.Server)
}
