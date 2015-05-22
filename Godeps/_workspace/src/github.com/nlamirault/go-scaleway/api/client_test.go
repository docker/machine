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
	//"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	//"net/url"
	"strings"
	"testing"
)

const (
	ScalewayUserID       = "12345678-520a-4ab7-9707-8bc1819a9e19"
	ScalewayToken        = "02468"
	ScalewayOrganization = "13579"
)

func loadJSON(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func getClient() *ScalewayClient {
	return NewClient(
		ScalewayToken,
		ScalewayUserID,
		ScalewayOrganization)
}

func newServer(content string) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			fmt.Printf("Req: %s", req.Method)
			if "DELETE" == req.Method {
				res.WriteHeader(http.StatusNoContent)
			}
			res.Header().Set(
				"Content-Type",
				"application/json; charset=utf-8")
			fmt.Fprintf(res, content)
		}))
}

func TestGettingUser(t *testing.T) {
	json, err := loadJSON("test_fixtures/user.json")
	if err != nil {
		t.Fatalf("Can't load JSON: %v", err)
	}
	ts := newServer(json)
	defer ts.Close()
	c := getClient()
	c.AccountURL = ts.URL
	response, err := c.GetUserInformations()
	// response, err := GetUserFromJSON(b)
	if err != nil {
		t.Fatalf("Can't decode json: %v", err)
	}
	if response.User.ID != "12345678-520a-4ab7-9707-8bc1819a9e19" {
		t.Fatalf("Invalid user id")
	}
	if response.User.Firstname != "Foo" {
		t.Fatalf("Invalud user firstname")
	}
	if response.User.Lastname != "Bar" {
		t.Fatalf("Invalid user lastname")
	}
}

func TestListUserOrganizations(t *testing.T) {
	json, err := loadJSON("test_fixtures/organizations.json")
	if err != nil {
		t.Fatalf("Can't load JSON: %v", err)
	}
	ts := newServer(json)
	defer ts.Close()
	c := getClient()
	c.AccountURL = ts.URL
	response, err := c.GetUserOrganizations()
	// response, err := GetOrganizationsFromJSON(b)
	if err != nil {
		t.Fatalf("Can't decode json: %v", err)
	}
	if len(response.Organizations) != 1 {
		t.Fatalf("Invalid number of organizations")
	}
	for _, org := range response.Organizations {
		fmt.Println(org)
		if org.ID != "19446e97-4a3b-4ccc-88f3-b65e3f31fb75" {
			t.Fatalf("Invalid organization id")
		}
		if org.Name != "foo.bar@gmail.com" {
			t.Fatalf("Invalid organization name")
		}
	}
}

func TestListUserTokens(t *testing.T) {
	json, err := loadJSON("test_fixtures/tokens.json")
	if err != nil {
		t.Fatalf("Can't load JSON: %v", err)
	}
	ts := newServer(json)
	defer ts.Close()
	c := getClient()
	c.AccountURL = ts.URL
	response, err := c.GetUserTokens()
	// response, err := GetTokensFromJSON(b)
	if err != nil {
		t.Fatalf("Can't decode json: %v", err)
	}
	for _, token := range response.Tokens {
		//fmt.Println(token)
		if token.UserID != "12345678-520a-4ab7-9707-8bc1819a9e19" {
			t.Fatalf("Invalid token userID")
		}
		if !strings.HasPrefix(token.ID, "13579") {
			t.Fatalf("Invalid token ID")
		}
	}
}

func TestGettingServer(t *testing.T) {
	json, err := loadJSON("test_fixtures/server.json")
	if err != nil {
		t.Fatalf("Can't load JSON: %v", err)
	}
	ts := newServer(json)
	defer ts.Close()
	c := getClient()
	c.ComputeURL = ts.URL
	response, err := c.GetServer("56e98092-6e05-4c89-9e76-b3610d38478c")
	//response, err := GetServerFromJSON(b)
	//fmt.Printf("Response: %v %v\n", response, err)
	if err != nil {
		t.Fatalf("Can't decode json: %v", err)
	}
	if response.Server.ID != "56e98092-6e05-4c89-9e76-b3610d38478c" {
		t.Fatalf("Invalid server id")
	}
	if response.Server.Name != "docker-lam" {
		t.Fatalf("Invalid server name")
	}
	if response.Server.State != "starting" {
		t.Fatalf("Invalid server state")
	}
	if response.Server.Organization != "19446e97-4a3b-4ccc-88f3-b65e3f31fb75" {
		t.Fatalf("Invalid server organization")
	}
}

func TestGettingServers(t *testing.T) {
	json, err := loadJSON("test_fixtures/servers.json")
	if err != nil {
		t.Fatalf("Can't load JSON: %v", err)
	}
	ts := newServer(json)
	defer ts.Close()
	c := getClient()
	c.ComputeURL = ts.URL
	response, err := c.GetServers()
	// response, err := GetServersFromJSON(b)
	//fmt.Printf("Response: %v %v\n", response, err)
	if err != nil {
		t.Fatalf("Can't decode json: %v", err)
	}
	for _, server := range response.Servers {
		//fmt.Println(server)
		if !strings.HasPrefix(server.ID, "02468") {
			t.Fatalf("Invalid server ID")
		}
		if server.Organization != "13579-4a3b-4ccc-88f3-b65e3f31fb75" {
			t.Fatalf("Invalid server organization")
		}
	}
}

func TestDeleteServer(t *testing.T) {
	json, err := loadJSON("test_fixtures/delete_server.json")
	if err != nil {
		t.Fatalf("Can't load JSON: %v", err)
	}
	ts := newServer(json)
	defer ts.Close()
	c := getClient()
	c.ComputeURL = ts.URL
	err = c.DeleteServer("56e98092-6e05-4c89-9e76-b3610d38478c")
	if err != nil {
		t.Fatalf("Invalid delete server response %v", err)
	}
}

func TestPoweroffServer(t *testing.T) {
	json, err := loadJSON("test_fixtures/server_poweroff.json")
	if err != nil {
		t.Fatalf("Can't load JSON: %v", err)
	}
	ts := newServer(json)
	defer ts.Close()
	c := getClient()
	c.ComputeURL = ts.URL
	response, err := c.PerformServerAction(
		"f5c94e15-1c11-4eab-a7a6-73db916b37c2",
		"poweroff")
	// response, _ := GetTaskFromJSON(b)
	if response.Task.Status != "pending" {
		t.Fatalf("Invalid task status")
	}
	if response.Task.Description != "server_poweroff" {
		t.Fatalf("Invalid task description")
	}
	if response.Task.Progress != 0 {
		t.Fatalf("Invalid task progress")
	}

}
