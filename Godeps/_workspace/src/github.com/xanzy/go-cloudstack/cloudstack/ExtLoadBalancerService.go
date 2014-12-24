//
// Copyright 2014, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package cloudstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type AddExternalLoadBalancerParams struct {
	p map[string]interface{}
}

func (p *AddExternalLoadBalancerParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["username"]; found {
		u.Set("username", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *AddExternalLoadBalancerParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

func (p *AddExternalLoadBalancerParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *AddExternalLoadBalancerParams) SetUsername(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["username"] = v
	return
}

func (p *AddExternalLoadBalancerParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new AddExternalLoadBalancerParams instance,
// as then you are sure you have configured all required params
func (s *ExtLoadBalancerService) NewAddExternalLoadBalancerParams(password string, url string, username string, zoneid string) *AddExternalLoadBalancerParams {
	p := &AddExternalLoadBalancerParams{}
	p.p = make(map[string]interface{})
	p.p["password"] = password
	p.p["url"] = url
	p.p["username"] = username
	p.p["zoneid"] = zoneid
	return p
}

// Adds F5 external load balancer appliance.
func (s *ExtLoadBalancerService) AddExternalLoadBalancer(p *AddExternalLoadBalancerParams) (*AddExternalLoadBalancerResponse, error) {
	resp, err := s.cs.newRequest("addExternalLoadBalancer", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddExternalLoadBalancerResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type AddExternalLoadBalancerResponse struct {
	Id               string `json:"id,omitempty"`
	Ipaddress        string `json:"ipaddress,omitempty"`
	Numretries       string `json:"numretries,omitempty"`
	Privateinterface string `json:"privateinterface,omitempty"`
	Publicinterface  string `json:"publicinterface,omitempty"`
	Username         string `json:"username,omitempty"`
	Zoneid           string `json:"zoneid,omitempty"`
}

type DeleteExternalLoadBalancerParams struct {
	p map[string]interface{}
}

func (p *DeleteExternalLoadBalancerParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteExternalLoadBalancerParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteExternalLoadBalancerParams instance,
// as then you are sure you have configured all required params
func (s *ExtLoadBalancerService) NewDeleteExternalLoadBalancerParams(id string) *DeleteExternalLoadBalancerParams {
	p := &DeleteExternalLoadBalancerParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a F5 external load balancer appliance added in a zone.
func (s *ExtLoadBalancerService) DeleteExternalLoadBalancer(p *DeleteExternalLoadBalancerParams) (*DeleteExternalLoadBalancerResponse, error) {
	resp, err := s.cs.newRequest("deleteExternalLoadBalancer", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteExternalLoadBalancerResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type DeleteExternalLoadBalancerResponse struct {
	Displaytext string `json:"displaytext,omitempty"`
	Success     string `json:"success,omitempty"`
}

type ListExternalLoadBalancersParams struct {
	p map[string]interface{}
}

func (p *ListExternalLoadBalancersParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["page"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("page", vv)
	}
	if v, found := p.p["pagesize"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("pagesize", vv)
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListExternalLoadBalancersParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListExternalLoadBalancersParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListExternalLoadBalancersParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListExternalLoadBalancersParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListExternalLoadBalancersParams instance,
// as then you are sure you have configured all required params
func (s *ExtLoadBalancerService) NewListExternalLoadBalancersParams() *ListExternalLoadBalancersParams {
	p := &ListExternalLoadBalancersParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *ExtLoadBalancerService) GetExternalLoadBalancerID(keyword string) (string, error) {
	p := &ListExternalLoadBalancersParams{}
	p.p = make(map[string]interface{})

	p.p["keyword"] = keyword

	l, err := s.ListExternalLoadBalancers(p)
	if err != nil {
		return "", err
	}

	if l.Count == 0 {
		return "", fmt.Errorf("No match found for %s: %+v", keyword, l)
	}

	if l.Count == 1 {
		return l.ExternalLoadBalancers[0].Id, nil
	}

	if l.Count > 1 {
		for _, v := range l.ExternalLoadBalancers {
			if v.Name == keyword {
				return v.Id, nil
			}
		}
	}
	return "", fmt.Errorf("Could not find an exact match for %s: %+v", keyword, l)
}

// Lists F5 external load balancer appliances added in a zone.
func (s *ExtLoadBalancerService) ListExternalLoadBalancers(p *ListExternalLoadBalancersParams) (*ListExternalLoadBalancersResponse, error) {
	resp, err := s.cs.newRequest("listExternalLoadBalancers", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListExternalLoadBalancersResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type ListExternalLoadBalancersResponse struct {
	Count                 int                     `json:"count"`
	ExternalLoadBalancers []*ExternalLoadBalancer `json:"externalloadbalancer"`
}

type ExternalLoadBalancer struct {
	Averageload             int    `json:"averageload,omitempty"`
	Capabilities            string `json:"capabilities,omitempty"`
	Clusterid               string `json:"clusterid,omitempty"`
	Clustername             string `json:"clustername,omitempty"`
	Clustertype             string `json:"clustertype,omitempty"`
	Cpuallocated            string `json:"cpuallocated,omitempty"`
	Cpunumber               int    `json:"cpunumber,omitempty"`
	Cpusockets              int    `json:"cpusockets,omitempty"`
	Cpuspeed                int    `json:"cpuspeed,omitempty"`
	Cpuused                 string `json:"cpuused,omitempty"`
	Cpuwithoverprovisioning string `json:"cpuwithoverprovisioning,omitempty"`
	Created                 string `json:"created,omitempty"`
	Disconnected            string `json:"disconnected,omitempty"`
	Disksizeallocated       int    `json:"disksizeallocated,omitempty"`
	Disksizetotal           int    `json:"disksizetotal,omitempty"`
	Events                  string `json:"events,omitempty"`
	Hahost                  bool   `json:"hahost,omitempty"`
	Hasenoughcapacity       bool   `json:"hasenoughcapacity,omitempty"`
	Hosttags                string `json:"hosttags,omitempty"`
	Hypervisor              string `json:"hypervisor,omitempty"`
	Hypervisorversion       string `json:"hypervisorversion,omitempty"`
	Id                      string `json:"id,omitempty"`
	Ipaddress               string `json:"ipaddress,omitempty"`
	Islocalstorageactive    bool   `json:"islocalstorageactive,omitempty"`
	Lastpinged              string `json:"lastpinged,omitempty"`
	Managementserverid      int    `json:"managementserverid,omitempty"`
	Memoryallocated         int    `json:"memoryallocated,omitempty"`
	Memorytotal             int    `json:"memorytotal,omitempty"`
	Memoryused              int    `json:"memoryused,omitempty"`
	Name                    string `json:"name,omitempty"`
	Networkkbsread          int    `json:"networkkbsread,omitempty"`
	Networkkbswrite         int    `json:"networkkbswrite,omitempty"`
	Oscategoryid            string `json:"oscategoryid,omitempty"`
	Oscategoryname          string `json:"oscategoryname,omitempty"`
	Podid                   string `json:"podid,omitempty"`
	Podname                 string `json:"podname,omitempty"`
	Removed                 string `json:"removed,omitempty"`
	Resourcestate           string `json:"resourcestate,omitempty"`
	State                   string `json:"state,omitempty"`
	Suitableformigration    bool   `json:"suitableformigration,omitempty"`
	Type                    string `json:"type,omitempty"`
	Version                 string `json:"version,omitempty"`
	Zoneid                  string `json:"zoneid,omitempty"`
	Zonename                string `json:"zonename,omitempty"`
}
