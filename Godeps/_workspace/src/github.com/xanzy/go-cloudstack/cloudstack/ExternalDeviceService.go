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
	"net/url"
	"strconv"
)

type DeleteCiscoNexusVSMParams struct {
	p map[string]interface{}
}

func (p *DeleteCiscoNexusVSMParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteCiscoNexusVSMParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteCiscoNexusVSMParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewDeleteCiscoNexusVSMParams(id string) *DeleteCiscoNexusVSMParams {
	p := &DeleteCiscoNexusVSMParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

//  delete a Cisco Nexus VSM device
func (s *ExternalDeviceService) DeleteCiscoNexusVSM(p *DeleteCiscoNexusVSMParams) (*DeleteCiscoNexusVSMResponse, error) {
	resp, err := s.cs.newRequest("deleteCiscoNexusVSM", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteCiscoNexusVSMResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	// If we have a async client, we need to wait for the async result
	if s.cs.async {
		b, warn, err := s.cs.GetAsyncJobResult(r.JobID, s.cs.timeout)
		if err != nil {
			return nil, err
		}
		// If 'warn' has a value it means the job is running longer than the configured
		// timeout, the resonse will contain the jobid of the running async job
		if warn != nil {
			return &r, warn
		}

		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}
	return &r, nil
}

type DeleteCiscoNexusVSMResponse struct {
	JobID       string `json:"jobid,omitempty"`
	Displaytext string `json:"displaytext,omitempty"`
	Success     bool   `json:"success,omitempty"`
}

type EnableCiscoNexusVSMParams struct {
	p map[string]interface{}
}

func (p *EnableCiscoNexusVSMParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *EnableCiscoNexusVSMParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new EnableCiscoNexusVSMParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewEnableCiscoNexusVSMParams(id string) *EnableCiscoNexusVSMParams {
	p := &EnableCiscoNexusVSMParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Enable a Cisco Nexus VSM device
func (s *ExternalDeviceService) EnableCiscoNexusVSM(p *EnableCiscoNexusVSMParams) (*EnableCiscoNexusVSMResponse, error) {
	resp, err := s.cs.newRequest("enableCiscoNexusVSM", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r EnableCiscoNexusVSMResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	// If we have a async client, we need to wait for the async result
	if s.cs.async {
		b, warn, err := s.cs.GetAsyncJobResult(r.JobID, s.cs.timeout)
		if err != nil {
			return nil, err
		}
		// If 'warn' has a value it means the job is running longer than the configured
		// timeout, the resonse will contain the jobid of the running async job
		if warn != nil {
			return &r, warn
		}

		b, err = getRawValue(b)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}
	return &r, nil
}

type EnableCiscoNexusVSMResponse struct {
	JobID            string `json:"jobid,omitempty"`
	Ipaddress        string `json:"ipaddress,omitempty"`
	Vsmconfigmode    string `json:"vsmconfigmode,omitempty"`
	Vsmconfigstate   string `json:"vsmconfigstate,omitempty"`
	Vsmctrlvlanid    int    `json:"vsmctrlvlanid,omitempty"`
	Vsmdeviceid      string `json:"vsmdeviceid,omitempty"`
	Vsmdevicename    string `json:"vsmdevicename,omitempty"`
	Vsmdevicestate   string `json:"vsmdevicestate,omitempty"`
	Vsmdomainid      string `json:"vsmdomainid,omitempty"`
	Vsmmgmtvlanid    string `json:"vsmmgmtvlanid,omitempty"`
	Vsmpktvlanid     int    `json:"vsmpktvlanid,omitempty"`
	Vsmstoragevlanid int    `json:"vsmstoragevlanid,omitempty"`
}

type DisableCiscoNexusVSMParams struct {
	p map[string]interface{}
}

func (p *DisableCiscoNexusVSMParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DisableCiscoNexusVSMParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DisableCiscoNexusVSMParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewDisableCiscoNexusVSMParams(id string) *DisableCiscoNexusVSMParams {
	p := &DisableCiscoNexusVSMParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// disable a Cisco Nexus VSM device
func (s *ExternalDeviceService) DisableCiscoNexusVSM(p *DisableCiscoNexusVSMParams) (*DisableCiscoNexusVSMResponse, error) {
	resp, err := s.cs.newRequest("disableCiscoNexusVSM", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DisableCiscoNexusVSMResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	// If we have a async client, we need to wait for the async result
	if s.cs.async {
		b, warn, err := s.cs.GetAsyncJobResult(r.JobID, s.cs.timeout)
		if err != nil {
			return nil, err
		}
		// If 'warn' has a value it means the job is running longer than the configured
		// timeout, the resonse will contain the jobid of the running async job
		if warn != nil {
			return &r, warn
		}

		b, err = getRawValue(b)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}
	return &r, nil
}

type DisableCiscoNexusVSMResponse struct {
	JobID            string `json:"jobid,omitempty"`
	Ipaddress        string `json:"ipaddress,omitempty"`
	Vsmconfigmode    string `json:"vsmconfigmode,omitempty"`
	Vsmconfigstate   string `json:"vsmconfigstate,omitempty"`
	Vsmctrlvlanid    int    `json:"vsmctrlvlanid,omitempty"`
	Vsmdeviceid      string `json:"vsmdeviceid,omitempty"`
	Vsmdevicename    string `json:"vsmdevicename,omitempty"`
	Vsmdevicestate   string `json:"vsmdevicestate,omitempty"`
	Vsmdomainid      string `json:"vsmdomainid,omitempty"`
	Vsmmgmtvlanid    string `json:"vsmmgmtvlanid,omitempty"`
	Vsmpktvlanid     int    `json:"vsmpktvlanid,omitempty"`
	Vsmstoragevlanid int    `json:"vsmstoragevlanid,omitempty"`
}

type ListCiscoNexusVSMsParams struct {
	p map[string]interface{}
}

func (p *ListCiscoNexusVSMsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["clusterid"]; found {
		u.Set("clusterid", v.(string))
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

func (p *ListCiscoNexusVSMsParams) SetClusterid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["clusterid"] = v
	return
}

func (p *ListCiscoNexusVSMsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListCiscoNexusVSMsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListCiscoNexusVSMsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListCiscoNexusVSMsParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListCiscoNexusVSMsParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewListCiscoNexusVSMsParams() *ListCiscoNexusVSMsParams {
	p := &ListCiscoNexusVSMsParams{}
	p.p = make(map[string]interface{})
	return p
}

// Retrieves a Cisco Nexus 1000v Virtual Switch Manager device associated with a Cluster
func (s *ExternalDeviceService) ListCiscoNexusVSMs(p *ListCiscoNexusVSMsParams) (*ListCiscoNexusVSMsResponse, error) {
	resp, err := s.cs.newRequest("listCiscoNexusVSMs", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListCiscoNexusVSMsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type ListCiscoNexusVSMsResponse struct {
	Count          int              `json:"count"`
	CiscoNexusVSMs []*CiscoNexusVSM `json:"cisconexusvsm"`
}

type CiscoNexusVSM struct {
	Ipaddress        string `json:"ipaddress,omitempty"`
	Vsmconfigmode    string `json:"vsmconfigmode,omitempty"`
	Vsmconfigstate   string `json:"vsmconfigstate,omitempty"`
	Vsmctrlvlanid    int    `json:"vsmctrlvlanid,omitempty"`
	Vsmdeviceid      string `json:"vsmdeviceid,omitempty"`
	Vsmdevicename    string `json:"vsmdevicename,omitempty"`
	Vsmdevicestate   string `json:"vsmdevicestate,omitempty"`
	Vsmdomainid      string `json:"vsmdomainid,omitempty"`
	Vsmmgmtvlanid    string `json:"vsmmgmtvlanid,omitempty"`
	Vsmpktvlanid     int    `json:"vsmpktvlanid,omitempty"`
	Vsmstoragevlanid int    `json:"vsmstoragevlanid,omitempty"`
}

type AddCiscoVnmcResourceParams struct {
	p map[string]interface{}
}

func (p *AddCiscoVnmcResourceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hostname"]; found {
		u.Set("hostname", v.(string))
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["username"]; found {
		u.Set("username", v.(string))
	}
	return u
}

func (p *AddCiscoVnmcResourceParams) SetHostname(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostname"] = v
	return
}

func (p *AddCiscoVnmcResourceParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

func (p *AddCiscoVnmcResourceParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *AddCiscoVnmcResourceParams) SetUsername(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["username"] = v
	return
}

// You should always use this function to get a new AddCiscoVnmcResourceParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewAddCiscoVnmcResourceParams(hostname string, password string, physicalnetworkid string, username string) *AddCiscoVnmcResourceParams {
	p := &AddCiscoVnmcResourceParams{}
	p.p = make(map[string]interface{})
	p.p["hostname"] = hostname
	p.p["password"] = password
	p.p["physicalnetworkid"] = physicalnetworkid
	p.p["username"] = username
	return p
}

// Adds a Cisco Vnmc Controller
func (s *ExternalDeviceService) AddCiscoVnmcResource(p *AddCiscoVnmcResourceParams) (*AddCiscoVnmcResourceResponse, error) {
	resp, err := s.cs.newRequest("addCiscoVnmcResource", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddCiscoVnmcResourceResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type AddCiscoVnmcResourceResponse struct {
}

type DeleteCiscoVnmcResourceParams struct {
	p map[string]interface{}
}

func (p *DeleteCiscoVnmcResourceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["resourceid"]; found {
		u.Set("resourceid", v.(string))
	}
	return u
}

func (p *DeleteCiscoVnmcResourceParams) SetResourceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["resourceid"] = v
	return
}

// You should always use this function to get a new DeleteCiscoVnmcResourceParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewDeleteCiscoVnmcResourceParams(resourceid string) *DeleteCiscoVnmcResourceParams {
	p := &DeleteCiscoVnmcResourceParams{}
	p.p = make(map[string]interface{})
	p.p["resourceid"] = resourceid
	return p
}

// Deletes a Cisco Vnmc controller
func (s *ExternalDeviceService) DeleteCiscoVnmcResource(p *DeleteCiscoVnmcResourceParams) (*DeleteCiscoVnmcResourceResponse, error) {
	resp, err := s.cs.newRequest("deleteCiscoVnmcResource", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteCiscoVnmcResourceResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type DeleteCiscoVnmcResourceResponse struct {
	Displaytext string `json:"displaytext,omitempty"`
	Success     string `json:"success,omitempty"`
}

type ListCiscoVnmcResourcesParams struct {
	p map[string]interface{}
}

func (p *ListCiscoVnmcResourcesParams) toURLValues() url.Values {
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
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["resourceid"]; found {
		u.Set("resourceid", v.(string))
	}
	return u
}

func (p *ListCiscoVnmcResourcesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListCiscoVnmcResourcesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListCiscoVnmcResourcesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListCiscoVnmcResourcesParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *ListCiscoVnmcResourcesParams) SetResourceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["resourceid"] = v
	return
}

// You should always use this function to get a new ListCiscoVnmcResourcesParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewListCiscoVnmcResourcesParams() *ListCiscoVnmcResourcesParams {
	p := &ListCiscoVnmcResourcesParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists Cisco VNMC controllers
func (s *ExternalDeviceService) ListCiscoVnmcResources(p *ListCiscoVnmcResourcesParams) (*ListCiscoVnmcResourcesResponse, error) {
	resp, err := s.cs.newRequest("listCiscoVnmcResources", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListCiscoVnmcResourcesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type ListCiscoVnmcResourcesResponse struct {
	Count              int                  `json:"count"`
	CiscoVnmcResources []*CiscoVnmcResource `json:"ciscovnmcresource"`
}

type CiscoVnmcResource struct {
}

type AddCiscoAsa1000vResourceParams struct {
	p map[string]interface{}
}

func (p *AddCiscoAsa1000vResourceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["clusterid"]; found {
		u.Set("clusterid", v.(string))
	}
	if v, found := p.p["hostname"]; found {
		u.Set("hostname", v.(string))
	}
	if v, found := p.p["insideportprofile"]; found {
		u.Set("insideportprofile", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	return u
}

func (p *AddCiscoAsa1000vResourceParams) SetClusterid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["clusterid"] = v
	return
}

func (p *AddCiscoAsa1000vResourceParams) SetHostname(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostname"] = v
	return
}

func (p *AddCiscoAsa1000vResourceParams) SetInsideportprofile(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["insideportprofile"] = v
	return
}

func (p *AddCiscoAsa1000vResourceParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

// You should always use this function to get a new AddCiscoAsa1000vResourceParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewAddCiscoAsa1000vResourceParams(clusterid string, hostname string, insideportprofile string, physicalnetworkid string) *AddCiscoAsa1000vResourceParams {
	p := &AddCiscoAsa1000vResourceParams{}
	p.p = make(map[string]interface{})
	p.p["clusterid"] = clusterid
	p.p["hostname"] = hostname
	p.p["insideportprofile"] = insideportprofile
	p.p["physicalnetworkid"] = physicalnetworkid
	return p
}

// Adds a Cisco Asa 1000v appliance
func (s *ExternalDeviceService) AddCiscoAsa1000vResource(p *AddCiscoAsa1000vResourceParams) (*AddCiscoAsa1000vResourceResponse, error) {
	resp, err := s.cs.newRequest("addCiscoAsa1000vResource", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddCiscoAsa1000vResourceResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type AddCiscoAsa1000vResourceResponse struct {
}

type DeleteCiscoAsa1000vResourceParams struct {
	p map[string]interface{}
}

func (p *DeleteCiscoAsa1000vResourceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["resourceid"]; found {
		u.Set("resourceid", v.(string))
	}
	return u
}

func (p *DeleteCiscoAsa1000vResourceParams) SetResourceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["resourceid"] = v
	return
}

// You should always use this function to get a new DeleteCiscoAsa1000vResourceParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewDeleteCiscoAsa1000vResourceParams(resourceid string) *DeleteCiscoAsa1000vResourceParams {
	p := &DeleteCiscoAsa1000vResourceParams{}
	p.p = make(map[string]interface{})
	p.p["resourceid"] = resourceid
	return p
}

// Deletes a Cisco ASA 1000v appliance
func (s *ExternalDeviceService) DeleteCiscoAsa1000vResource(p *DeleteCiscoAsa1000vResourceParams) (*DeleteCiscoAsa1000vResourceResponse, error) {
	resp, err := s.cs.newRequest("deleteCiscoAsa1000vResource", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteCiscoAsa1000vResourceResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type DeleteCiscoAsa1000vResourceResponse struct {
	Displaytext string `json:"displaytext,omitempty"`
	Success     string `json:"success,omitempty"`
}

type ListCiscoAsa1000vResourcesParams struct {
	p map[string]interface{}
}

func (p *ListCiscoAsa1000vResourcesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hostname"]; found {
		u.Set("hostname", v.(string))
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
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["resourceid"]; found {
		u.Set("resourceid", v.(string))
	}
	return u
}

func (p *ListCiscoAsa1000vResourcesParams) SetHostname(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostname"] = v
	return
}

func (p *ListCiscoAsa1000vResourcesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListCiscoAsa1000vResourcesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListCiscoAsa1000vResourcesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListCiscoAsa1000vResourcesParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *ListCiscoAsa1000vResourcesParams) SetResourceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["resourceid"] = v
	return
}

// You should always use this function to get a new ListCiscoAsa1000vResourcesParams instance,
// as then you are sure you have configured all required params
func (s *ExternalDeviceService) NewListCiscoAsa1000vResourcesParams() *ListCiscoAsa1000vResourcesParams {
	p := &ListCiscoAsa1000vResourcesParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists Cisco ASA 1000v appliances
func (s *ExternalDeviceService) ListCiscoAsa1000vResources(p *ListCiscoAsa1000vResourcesParams) (*ListCiscoAsa1000vResourcesResponse, error) {
	resp, err := s.cs.newRequest("listCiscoAsa1000vResources", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListCiscoAsa1000vResourcesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type ListCiscoAsa1000vResourcesResponse struct {
	Count                  int                      `json:"count"`
	CiscoAsa1000vResources []*CiscoAsa1000vResource `json:"ciscoasa1000vresource"`
}

type CiscoAsa1000vResource struct {
}
