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
	//"fmt"

	//log "github.com/Sirupsen/logrus"
	"github.com/nlamirault/go-scaleway/log"
)

// Token represents an identifier associated with your account.
// It is used to authenticate commands in the APIs
type Token struct {
	ID       string `json:"id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Creation string `json:"creation_date,omitempty"`
	Expires  string `json:"expires,omitempty"`
}

// TokenResponse represents JSON response of server for a token
type TokenResponse struct {
	Token Token `json:"token,omitempty"`
}

// TokensResponse represents JSON response of server for tokens
type TokensResponse struct {
	Tokens []Token
}

// Display log the Token caracteristics
func (t Token) Display() {
	log.Infof("Id        : %s", t.ID)
	log.Infof("UserId    : %s", t.UserID)
	log.Infof("Creation  : %s", t.Creation)
	log.Infof("Expires   : %s", t.Expires)
}
