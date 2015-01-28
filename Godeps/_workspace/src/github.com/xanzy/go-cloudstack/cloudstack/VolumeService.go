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
	"strings"
)

type AttachVolumeParams struct {
	p map[string]interface{}
}

func (p *AttachVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["deviceid"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("deviceid", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *AttachVolumeParams) SetDeviceid(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["deviceid"] = v
	return
}

func (p *AttachVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *AttachVolumeParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new AttachVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewAttachVolumeParams(id string, virtualmachineid string) *AttachVolumeParams {
	p := &AttachVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Attaches a disk volume to a virtual machine.
func (s *VolumeService) AttachVolume(p *AttachVolumeParams) (*AttachVolumeResponse, error) {
	resp, err := s.cs.newRequest("attachVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AttachVolumeResponse
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

type AttachVolumeResponse struct {
	JobID                      string `json:"jobid,omitempty"`
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type UploadVolumeParams struct {
	p map[string]interface{}
}

func (p *UploadVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["checksum"]; found {
		u.Set("checksum", v.(string))
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["format"]; found {
		u.Set("format", v.(string))
	}
	if v, found := p.p["imagestoreuuid"]; found {
		u.Set("imagestoreuuid", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *UploadVolumeParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *UploadVolumeParams) SetChecksum(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["checksum"] = v
	return
}

func (p *UploadVolumeParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *UploadVolumeParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *UploadVolumeParams) SetFormat(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["format"] = v
	return
}

func (p *UploadVolumeParams) SetImagestoreuuid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["imagestoreuuid"] = v
	return
}

func (p *UploadVolumeParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UploadVolumeParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *UploadVolumeParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *UploadVolumeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new UploadVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewUploadVolumeParams(format string, name string, url string, zoneid string) *UploadVolumeParams {
	p := &UploadVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["format"] = format
	p.p["name"] = name
	p.p["url"] = url
	p.p["zoneid"] = zoneid
	return p
}

// Uploads a data disk.
func (s *VolumeService) UploadVolume(p *UploadVolumeParams) (*UploadVolumeResponse, error) {
	resp, err := s.cs.newRequest("uploadVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UploadVolumeResponse
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

type UploadVolumeResponse struct {
	JobID                      string `json:"jobid,omitempty"`
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type DetachVolumeParams struct {
	p map[string]interface{}
}

func (p *DetachVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["deviceid"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("deviceid", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *DetachVolumeParams) SetDeviceid(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["deviceid"] = v
	return
}

func (p *DetachVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *DetachVolumeParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new DetachVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewDetachVolumeParams() *DetachVolumeParams {
	p := &DetachVolumeParams{}
	p.p = make(map[string]interface{})
	return p
}

// Detaches a disk volume from a virtual machine.
func (s *VolumeService) DetachVolume(p *DetachVolumeParams) (*DetachVolumeResponse, error) {
	resp, err := s.cs.newRequest("detachVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DetachVolumeResponse
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

type DetachVolumeResponse struct {
	JobID                      string `json:"jobid,omitempty"`
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type CreateVolumeParams struct {
	p map[string]interface{}
}

func (p *CreateVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["displayvolume"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvolume", vv)
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["maxiops"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("maxiops", vv)
	}
	if v, found := p.p["miniops"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("miniops", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["size"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("size", vv)
	}
	if v, found := p.p["snapshotid"]; found {
		u.Set("snapshotid", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *CreateVolumeParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *CreateVolumeParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *CreateVolumeParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *CreateVolumeParams) SetDisplayvolume(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvolume"] = v
	return
}

func (p *CreateVolumeParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreateVolumeParams) SetMaxiops(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["maxiops"] = v
	return
}

func (p *CreateVolumeParams) SetMiniops(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["miniops"] = v
	return
}

func (p *CreateVolumeParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateVolumeParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *CreateVolumeParams) SetSize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["size"] = v
	return
}

func (p *CreateVolumeParams) SetSnapshotid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["snapshotid"] = v
	return
}

func (p *CreateVolumeParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

func (p *CreateVolumeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new CreateVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewCreateVolumeParams(name string) *CreateVolumeParams {
	p := &CreateVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	return p
}

// Creates a disk volume from a disk offering. This disk volume must still be attached to a virtual machine to make use of it.
func (s *VolumeService) CreateVolume(p *CreateVolumeParams) (*CreateVolumeResponse, error) {
	resp, err := s.cs.newRequest("createVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateVolumeResponse
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

type CreateVolumeResponse struct {
	JobID                      string `json:"jobid,omitempty"`
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type DeleteVolumeParams struct {
	p map[string]interface{}
}

func (p *DeleteVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewDeleteVolumeParams(id string) *DeleteVolumeParams {
	p := &DeleteVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a detached disk volume.
func (s *VolumeService) DeleteVolume(p *DeleteVolumeParams) (*DeleteVolumeResponse, error) {
	resp, err := s.cs.newRequest("deleteVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteVolumeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type DeleteVolumeResponse struct {
	Displaytext string `json:"displaytext,omitempty"`
	Success     string `json:"success,omitempty"`
}

type ListVolumesParams struct {
	p map[string]interface{}
}

func (p *ListVolumesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["displayvolume"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvolume", vv)
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["isrecursive"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("isrecursive", vv)
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["listall"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("listall", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["page"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("page", vv)
	}
	if v, found := p.p["pagesize"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("pagesize", vv)
	}
	if v, found := p.p["podid"]; found {
		u.Set("podid", v.(string))
	}
	if v, found := p.p["podid"]; found {
		u.Set("podid", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["tags"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("tags[%d].key", i), k)
			u.Set(fmt.Sprintf("tags[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListVolumesParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListVolumesParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *ListVolumesParams) SetDisplayvolume(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvolume"] = v
	return
}

func (p *ListVolumesParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListVolumesParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *ListVolumesParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListVolumesParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListVolumesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListVolumesParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListVolumesParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListVolumesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListVolumesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListVolumesParams) SetPodid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["podid"] = v
	return
}

func (p *ListVolumesParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListVolumesParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *ListVolumesParams) SetTags(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *ListVolumesParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["volumeType"] = v
	return
}

func (p *ListVolumesParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

func (p *ListVolumesParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListVolumesParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewListVolumesParams() *ListVolumesParams {
	p := &ListVolumesParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VolumeService) GetVolumeID(name string) (string, error) {
	p := &ListVolumesParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	l, err := s.ListVolumes(p)
	if err != nil {
		return "", err
	}

	if l.Count == 0 {
		return "", fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.Volumes[0].Id, nil
	}

	if l.Count > 1 {
		for _, v := range l.Volumes {
			if v.Name == name {
				return v.Id, nil
			}
		}
	}
	return "", fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VolumeService) GetVolumeByName(name string) (*Volume, int, error) {
	id, err := s.GetVolumeID(name)
	if err != nil {
		return nil, -1, err
	}

	r, count, err := s.GetVolumeByID(id)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VolumeService) GetVolumeByID(id string) (*Volume, int, error) {
	p := &ListVolumesParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	l, err := s.ListVolumes(p)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", id)) {
			return nil, 0, fmt.Errorf("No match found for %s: %+v", id, l)
		}
		return nil, -1, err
	}

	if l.Count == 0 {
		return nil, l.Count, fmt.Errorf("No match found for %s: %+v", id, l)
	}

	if l.Count == 1 {
		return l.Volumes[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for Volume UUID: %s!", id)
}

// Lists all volumes.
func (s *VolumeService) ListVolumes(p *ListVolumesParams) (*ListVolumesResponse, error) {
	resp, err := s.cs.newRequest("listVolumes", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListVolumesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type ListVolumesResponse struct {
	Count   int       `json:"count"`
	Volumes []*Volume `json:"volume"`
}

type Volume struct {
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type ExtractVolumeParams struct {
	p map[string]interface{}
}

func (p *ExtractVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["mode"]; found {
		u.Set("mode", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ExtractVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ExtractVolumeParams) SetMode(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["mode"] = v
	return
}

func (p *ExtractVolumeParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *ExtractVolumeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ExtractVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewExtractVolumeParams(id string, mode string, zoneid string) *ExtractVolumeParams {
	p := &ExtractVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	p.p["mode"] = mode
	p.p["zoneid"] = zoneid
	return p
}

// Extracts volume
func (s *VolumeService) ExtractVolume(p *ExtractVolumeParams) (*ExtractVolumeResponse, error) {
	resp, err := s.cs.newRequest("extractVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ExtractVolumeResponse
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

type ExtractVolumeResponse struct {
	JobID            string `json:"jobid,omitempty"`
	Accountid        string `json:"accountid,omitempty"`
	Created          string `json:"created,omitempty"`
	ExtractId        string `json:"extractId,omitempty"`
	ExtractMode      string `json:"extractMode,omitempty"`
	Id               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	Resultstring     string `json:"resultstring,omitempty"`
	State            string `json:"state,omitempty"`
	Status           string `json:"status,omitempty"`
	Storagetype      string `json:"storagetype,omitempty"`
	Uploadpercentage int    `json:"uploadpercentage,omitempty"`
	Url              string `json:"url,omitempty"`
	Zoneid           string `json:"zoneid,omitempty"`
	Zonename         string `json:"zonename,omitempty"`
}

type MigrateVolumeParams struct {
	p map[string]interface{}
}

func (p *MigrateVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["livemigrate"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("livemigrate", vv)
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["volumeid"]; found {
		u.Set("volumeid", v.(string))
	}
	return u
}

func (p *MigrateVolumeParams) SetLivemigrate(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["livemigrate"] = v
	return
}

func (p *MigrateVolumeParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *MigrateVolumeParams) SetVolumeid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["volumeid"] = v
	return
}

// You should always use this function to get a new MigrateVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewMigrateVolumeParams(storageid string, volumeid string) *MigrateVolumeParams {
	p := &MigrateVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["storageid"] = storageid
	p.p["volumeid"] = volumeid
	return p
}

// Migrate volume
func (s *VolumeService) MigrateVolume(p *MigrateVolumeParams) (*MigrateVolumeResponse, error) {
	resp, err := s.cs.newRequest("migrateVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r MigrateVolumeResponse
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

type MigrateVolumeResponse struct {
	JobID                      string `json:"jobid,omitempty"`
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type ResizeVolumeParams struct {
	p map[string]interface{}
}

func (p *ResizeVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["shrinkok"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("shrinkok", vv)
	}
	if v, found := p.p["size"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("size", vv)
	}
	return u
}

func (p *ResizeVolumeParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *ResizeVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ResizeVolumeParams) SetShrinkok(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["shrinkok"] = v
	return
}

func (p *ResizeVolumeParams) SetSize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["size"] = v
	return
}

// You should always use this function to get a new ResizeVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewResizeVolumeParams(id string) *ResizeVolumeParams {
	p := &ResizeVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Resizes a volume
func (s *VolumeService) ResizeVolume(p *ResizeVolumeParams) (*ResizeVolumeResponse, error) {
	resp, err := s.cs.newRequest("resizeVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ResizeVolumeResponse
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

type ResizeVolumeResponse struct {
	JobID                      string `json:"jobid,omitempty"`
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}

type UpdateVolumeParams struct {
	p map[string]interface{}
}

func (p *UpdateVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["chaininfo"]; found {
		u.Set("chaininfo", v.(string))
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["displayvolume"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvolume", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["path"]; found {
		u.Set("path", v.(string))
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	return u
}

func (p *UpdateVolumeParams) SetChaininfo(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["chaininfo"] = v
	return
}

func (p *UpdateVolumeParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *UpdateVolumeParams) SetDisplayvolume(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvolume"] = v
	return
}

func (p *UpdateVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateVolumeParams) SetPath(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["path"] = v
	return
}

func (p *UpdateVolumeParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *UpdateVolumeParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

// You should always use this function to get a new UpdateVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewUpdateVolumeParams() *UpdateVolumeParams {
	p := &UpdateVolumeParams{}
	p.p = make(map[string]interface{})
	return p
}

// Updates the volume.
func (s *VolumeService) UpdateVolume(p *UpdateVolumeParams) (*UpdateVolumeResponse, error) {
	resp, err := s.cs.newRequest("updateVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateVolumeResponse
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

type UpdateVolumeResponse struct {
	JobID                      string `json:"jobid,omitempty"`
	Account                    string `json:"account,omitempty"`
	Attached                   string `json:"attached,omitempty"`
	Chaininfo                  string `json:"chaininfo,omitempty"`
	Created                    string `json:"created,omitempty"`
	Destroyed                  bool   `json:"destroyed,omitempty"`
	Deviceid                   int    `json:"deviceid,omitempty"`
	DiskBytesReadRate          int    `json:"diskBytesReadRate,omitempty"`
	DiskBytesWriteRate         int    `json:"diskBytesWriteRate,omitempty"`
	DiskIopsReadRate           int    `json:"diskIopsReadRate,omitempty"`
	DiskIopsWriteRate          int    `json:"diskIopsWriteRate,omitempty"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext,omitempty"`
	Diskofferingid             string `json:"diskofferingid,omitempty"`
	Diskofferingname           string `json:"diskofferingname,omitempty"`
	Displayvolume              bool   `json:"displayvolume,omitempty"`
	Domain                     string `json:"domain,omitempty"`
	Domainid                   string `json:"domainid,omitempty"`
	Hypervisor                 string `json:"hypervisor,omitempty"`
	Id                         string `json:"id,omitempty"`
	Isextractable              bool   `json:"isextractable,omitempty"`
	Isodisplaytext             string `json:"isodisplaytext,omitempty"`
	Isoid                      string `json:"isoid,omitempty"`
	Isoname                    string `json:"isoname,omitempty"`
	Maxiops                    int    `json:"maxiops,omitempty"`
	Miniops                    int    `json:"miniops,omitempty"`
	Name                       string `json:"name,omitempty"`
	Path                       string `json:"path,omitempty"`
	Project                    string `json:"project,omitempty"`
	Projectid                  string `json:"projectid,omitempty"`
	Quiescevm                  bool   `json:"quiescevm,omitempty"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext,omitempty"`
	Serviceofferingid          string `json:"serviceofferingid,omitempty"`
	Serviceofferingname        string `json:"serviceofferingname,omitempty"`
	Size                       int    `json:"size,omitempty"`
	Snapshotid                 string `json:"snapshotid,omitempty"`
	State                      string `json:"state,omitempty"`
	Status                     string `json:"status,omitempty"`
	Storage                    string `json:"storage,omitempty"`
	Storageid                  string `json:"storageid,omitempty"`
	Storagetype                string `json:"storagetype,omitempty"`
	Tags                       []struct {
		Account      string `json:"account,omitempty"`
		Customer     string `json:"customer,omitempty"`
		Domain       string `json:"domain,omitempty"`
		Domainid     string `json:"domainid,omitempty"`
		Key          string `json:"key,omitempty"`
		Project      string `json:"project,omitempty"`
		Projectid    string `json:"projectid,omitempty"`
		Resourceid   string `json:"resourceid,omitempty"`
		Resourcetype string `json:"resourcetype,omitempty"`
		Value        string `json:"value,omitempty"`
	} `json:"tags,omitempty"`
	Templatedisplaytext string `json:"templatedisplaytext,omitempty"`
	Templateid          string `json:"templateid,omitempty"`
	Templatename        string `json:"templatename,omitempty"`
	Type                string `json:"type,omitempty"`
	Virtualmachineid    string `json:"virtualmachineid,omitempty"`
	Vmdisplayname       string `json:"vmdisplayname,omitempty"`
	Vmname              string `json:"vmname,omitempty"`
	Vmstate             string `json:"vmstate,omitempty"`
	Zoneid              string `json:"zoneid,omitempty"`
	Zonename            string `json:"zonename,omitempty"`
}
