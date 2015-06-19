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

	//log "github.com/Sirupsen/logrus"
	"github.com/nlamirault/go-scaleway/log"
)

// Image represents the image entity
type Image struct {
	ID           string `json:"id,omitempty"`
	Arch         string `json:"arch,omitempty"`
	Name         string `json:"name,omitempty"`
	Creation     string `json:"creation_date,omitempty"`
	Modification string `json:"modification_date,omitempty"`
	Organization string `json:"organization,omitempty"`
	Public       bool   `json:"public,omitempty"`
}

// ImageResponse represents JSON response of an image
type ImageResponse struct {
	Image Image `json:"image,omitempty"`
}

// ImagesResponse represents a list of volumes in JSON
type ImagesResponse struct {
	Images []Image
}

// Display log the Image caracteristics
func (i Image) Display() {
	log.Infof("Id            : %s", i.ID)
	log.Infof("Name          : %s", i.Name)
	log.Infof("Arch          : %s", i.Arch)
	log.Infof("Organisation  : %s", i.Organization)
	log.Infof("Creation      : %s", i.Creation)
	log.Infof("Modificaton   : %s", i.Modification)
}
