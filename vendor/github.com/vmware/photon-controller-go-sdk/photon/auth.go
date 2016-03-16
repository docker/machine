// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package photon

import (
	"encoding/json"
	"strings"

	"github.com/vmware/photon-controller-go-sdk/photon/internal/rest"
)

// Contains functionality for auth API.
type AuthAPI struct {
	client *Client
}

var authUrl string = "/auth"
var tokenUrl string = "/openidconnect/token"

// Gets authentication info.
func (api *AuthAPI) Get() (info *AuthInfo, err error) {
	res, err := rest.Get(api.client.httpClient, api.client.Endpoint+authUrl, "")
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	info = &AuthInfo{}
	err = json.NewDecoder(res.Body).Decode(info)
	return
}

// Gets Tokens from username/password.
func (api *AuthAPI) GetTokensByPassword(username string, password string) (tokenOptions *TokenOptions, err error) {
	body := strings.NewReader("grant_type=password&username=" + username + "&password=" + password + "&scope=openid offline_access")
	res, err := rest.Post(api.client.httpClient,
		api.client.AuthEndpoint+tokenUrl,
		"application/x-www-form-urlencoded",
		body,
		"")
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	tokenOptions = &TokenOptions{}
	err = json.NewDecoder(res.Body).Decode(tokenOptions)
	return
}

// Gets tokens from refresh token.
func (api *AuthAPI) GetTokensByRefreshToken(refreshtoken string) (tokenOptions *TokenOptions, err error) {
	body := strings.NewReader("grant_type=prefresh_token&refresh_token=" + refreshtoken + "&scope=openid offline_access")
	res, err := rest.Post(api.client.httpClient,
		api.client.AuthEndpoint+tokenUrl,
		"application/x-www-form-urlencoded",
		body,
		"")
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	tokenOptions = &TokenOptions{}
	err = json.NewDecoder(res.Body).Decode(tokenOptions)
	return
}
