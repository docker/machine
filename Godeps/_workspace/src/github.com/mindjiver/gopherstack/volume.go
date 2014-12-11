package gopherstack

import (
	"net/url"
)

// List volumes for Virtual Machine by it's ID
func (c CloudstackClient) ListVolumes(vmid string) (ListVolumesResponse, error) {
	var resp ListVolumesResponse
	params := url.Values{}
	params.Set("virtualmachineid", vmid)
	params.Set("listall", "true")
	response, err := NewRequest(c, "listVolumes", params)
	if err != nil {
		return resp, err
	}

	resp = response.(ListVolumesResponse)

	return resp, err
}

type Volume struct {
	Account                    string        `json:"account"`
	Created                    string        `json:"created"`
	Destroyed                  bool          `json:"destroyed"`
	Deviceid                   float64       `json:"deviceid"`
	Domain                     string        `json:"domain"`
	Domainid                   string        `json:"domainid"`
	ID                         string        `json:"id"`
	Isextractable              bool          `json:"isextractable"`
	Name                       string        `json:"name"`
	Serviceofferingdisplaytext string        `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string        `json:"serviceofferingid"`
	Serviceofferingname        string        `json:"serviceofferingname"`
	Size                       float64       `json:"size"`
	State                      string        `json:"state"`
	Storage                    string        `json:"storage"`
	Storagetype                string        `json:"storagetype"`
	Tags                       []interface{} `json:"tags"`
	Type                       string        `json:"type"`
	Virtualmachineid           string        `json:"virtualmachineid"`
	Vmdisplayname              string        `json:"vmdisplayname"`
	Vmname                     string        `json:"vmname"`
	Vmstate                    string        `json:"vmstate"`
	Zoneid                     string        `json:"zoneid"`
	Zonename                   string        `json:"zonename"`
}

type ListVolumesResponse struct {
	Listvolumesresponse struct {
		Count  float64  `json:"count"`
		Volume []Volume `json:"volume"`
	} `json:"listvolumesresponse"`
}
