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
	"io/ioutil"
	"net/http"
	"strings"

	//log "github.com/Sirupsen/logrus"
	"github.com/nlamirault/go-scaleway/log"
)

const (
	computeURL = "https://api.scaleway.com"
	accountURL = "https://account.scaleway.com"
)

// ScalewayClient is a client for the Scaleway API.
// Token is to authenticate to the API
// UserID represents your user identifiant
// Organization is the ID of the user's organization
type ScalewayClient struct {
	Token        string
	UserID       string
	Organization string
	Client       *http.Client
	ComputeURL   string
	AccountURL   string
}

// NewClient creates a new Scaleway API client using API token.
// userid can be an empty string - defaults to the token's user id
// organization can be an empty string - defaults to the user's primary organization
func NewClient(token string, userid string, organization string) *ScalewayClient {
	log.Debugf("Creating client using token=%s userid=%s org=%s", token, userid, organization)
	client := &ScalewayClient{
		Token:        token,
		UserID:       userid,
		Organization: organization,
		Client:       &http.Client{},
		ComputeURL:   computeURL,
		AccountURL:   accountURL,
	}
	return client
}

// GetUserInformations list informations about your user account
func (c ScalewayClient) GetUserInformations() (UserResponse, error) {
	var data UserResponse
	if err := c.SetUserFromToken(); err != nil {
		return data, err
	}
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/users/%s", c.AccountURL, c.UserID),
		&data)
	return data, err
}

// SetUserFromToken set UserID from Token if left empty
func (c ScalewayClient) SetUserFromToken() error {
	if c.UserID != "" {
		return nil
	}
	response, err := c.GetUserToken(c.Token)
	if err != nil {
		return err
	}
	c.UserID = response.Token.UserID
	return nil
}

// SetOrganizationFromToken set Organization from Token if left empty
func (c ScalewayClient) SetOrganizationFromToken() error {
	if c.Organization != "" {
		return nil
	}
	response, err := c.GetUserOrganizations()
	if err != nil {
		return err
	}
	if len(response.Organizations) > 0 {
		c.Organization = response.Organizations[0].ID
	}
	return nil
}

// GetUserOrganizations list all organizations associate with your account
func (c ScalewayClient) GetUserOrganizations() (OrganizationsResponse, error) {
	var data OrganizationsResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/organizations", c.AccountURL),
		&data)
	return data, err
}

// GetUserTokens list all tokens associate with your account
func (c ScalewayClient) GetUserTokens() (TokensResponse, error) {
	var data TokensResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/tokens", c.AccountURL),
		&data)
	return data, err
}

//GetUserToken lList an individual Token
func (c ScalewayClient) GetUserToken(tokenID string) (TokenResponse, error) {
	var data TokenResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/tokens/%s", c.AccountURL, tokenID),
		&data)
	return data, err
}

// CreateToken authenticates a user against their email, password,
// and then returns a new Token, which can be used until it expires.
// email is the user email
// password is the user password
// expires is if you want a token wich expires or not
func (c ScalewayClient) CreateToken(email string, password string, expires bool) (TokenResponse, error) {
	var data TokenResponse
	json := fmt.Sprintf(`{"email": "%s", "password": "%s", "expires": %t}`,
		email, password, expires)
	err := postAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/tokens", c.AccountURL),
		[]byte(json),
		&data)
	return data, err
}

// DeleteToken delete a specific token
// tokenID is the token unique identifier
func (c ScalewayClient) DeleteToken(tokenID string) error {
	return deleteAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/tokens/%s", c.AccountURL, tokenID),
		nil)
}

// UpdateToken increase Token expiration time of 30 minutes
// tokenID is the token unique identifier
func (c ScalewayClient) UpdateToken(tokenID string) (TokenResponse, error) {
	var data TokenResponse
	json := `{"expires": true}`
	err := patchAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/tokens/%s", c.AccountURL, tokenID),
		[]byte(json),
		&data)
	return data, err
}

// CreateServer creates a new server
// name is the server name
// image is the image unique identifier
func (c ScalewayClient) CreateServer(name string, image string) (ServerResponse, error) {
	var data ServerResponse
	if err := c.SetOrganizationFromToken(); err != nil {
		return data, err
	}
	json := fmt.Sprintf(`{"name": "%s", "organization": "%s", "image": "%s", "tags": ["docker-machine"]}`,
		name, c.Organization, image)
	err := postAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/servers", c.ComputeURL),
		[]byte(json),
		&data)
	return data, err
}

// ListServerActions list actions to be applied on a server
// serverID is the server unique identifier
func (c ScalewayClient) ListServerActions(serverID string) (ServerActionsResponse, error) {
	var data ServerActionsResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/servers/%s/action", c.ComputeURL, serverID),
		&data)
	return data, err
}

// DeleteServer delete a specific server
// serverID is the server unique identifier
func (c ScalewayClient) DeleteServer(serverID string) error {
	return deleteAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/servers/%s", c.ComputeURL, serverID),
		nil)
}

// GetServer list an individual server
// serverID is the server unique identifier
func (c ScalewayClient) GetServer(serverID string) (ServerResponse, error) {
	var data ServerResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/servers/%s", c.ComputeURL, serverID),
		&data)
	return data, err
}

// GetServers list all servers associate with your account
func (c ScalewayClient) GetServers() (ServersResponse, error) {
	var data ServersResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/servers", c.ComputeURL),
		&data)
	return data, err
}

// PerformServerAction execute an action on a server
// serverID is the server unique identifier
// action is the action to execute
func (c ScalewayClient) PerformServerAction(serverID string, action string) (TaskResponse, error) {
	var data TaskResponse
	json := fmt.Sprintf(`{"action": "%s"}`, action)
	err := postAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/servers/%s/action", c.ComputeURL, serverID),
		[]byte(json),
		&data)
	return data, err
}

// GetVolume list an individual volume
// volumeID is the volume unique identifier
func (c ScalewayClient) GetVolume(volumeID string) (VolumeResponse, error) {
	var data VolumeResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/volumes/%s", c.ComputeURL, volumeID),
		&data)
	return data, err
}

// DeleteVolume delete a specific volume
// volumeID is the volume unique identifier
func (c ScalewayClient) DeleteVolume(volumeID string) error {
	return deleteAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/volumes/%s", c.ComputeURL, volumeID),
		nil)
}

// CreateVolume creates a new volume
// name is the volume name
// volumeType is the volume type
// size is the volume size
func (c ScalewayClient) CreateVolume(name string, volumeType string, size int) (VolumeResponse, error) {
	var data VolumeResponse
	if err := c.SetOrganizationFromToken(); err != nil {
		return data, err
	}
	json := fmt.Sprintf(`{"name": "%s", "organization": "%s", "volume_type": "%s", "size": %d}`,
		name, c.Organization, volumeType, size)
	err := postAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/volumes", c.ComputeURL),
		[]byte(json),
		&data)
	return data, err
}

// GetVolumes list all volumes associate with your account
func (c ScalewayClient) GetVolumes() (VolumesResponse, error) {
	var data VolumesResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/volumes", c.ComputeURL),
		&data)
	return data, err
}

// GetImages list all images associate with your account
func (c ScalewayClient) GetImages() (ImagesResponse, error) {
	var data ImagesResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/images", c.ComputeURL),
		&data)
	return data, err
}

// GetImage list an individual image
// volumeID is the image unique identifier
func (c ScalewayClient) GetImage(volumeID string) (ImageResponse, error) {
	var data ImageResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/images/%s", c.ComputeURL, volumeID),
		&data)
	return data, err
}

// DeleteImage delete a specific volume
// volumeID is the volume unique identifier
func (c ScalewayClient) DeleteImage(imageID string) error {
	return deleteAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/images/%s", c.ComputeURL, imageID),
		nil)
}

// UploadPublicKey update user SSH keys
// keyPath is the complete path of the SSH key
func (c ScalewayClient) UploadPublicKey(keyPath string) (UserResponse, error) {
	var data UserResponse
	if err := c.SetUserFromToken(); err != nil {
		return data, err
	}
	publicKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return data, err
	}
	json := fmt.Sprintf(`{"ssh_public_keys": [{"key": "%s"}]}`,
		strings.TrimSpace(string(publicKey)))
	err = patchAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/users/%s", c.AccountURL, c.UserID),
		[]byte(json),
		&data)
	return data, err
}

// CreateIP creates a new reserved IP address
func (c ScalewayClient) CreateIP() (IPAddressResponse, error) {
	var data IPAddressResponse
	if err := c.SetOrganizationFromToken(); err != nil {
		return data, err
	}
	json := fmt.Sprintf(`{"organization": "%s"}`, c.Organization)
	err := postAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/ips", c.ComputeURL),
		[]byte(json),
		&data)
	return data, err
}

// GetIPs list all IPs associate with your account
func (c ScalewayClient) GetIPs() (IPAddressesResponse, error) {
	var data IPAddressesResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/ips", c.ComputeURL),
		&data)
	return data, err
}

// GetIP list an individual IP address
// ipID is the IP unique identifier
func (c ScalewayClient) GetIP(ipID string) (IPAddressResponse, error) {
	var data IPAddressResponse
	err := getAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/ips/%s", c.ComputeURL, ipID),
		&data)
	return data, err
}

// DeleteIP delete a specific IP address
// ipID is the IP unique identifier
func (c ScalewayClient) DeleteIP(ipID string) error {
	return deleteAPIResource(
		c.Client,
		c.Token,
		fmt.Sprintf("%s/ips/%s", c.ComputeURL, ipID),
		nil)
}
