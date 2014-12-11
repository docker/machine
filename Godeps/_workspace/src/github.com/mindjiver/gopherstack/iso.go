package gopherstack

import (
	"net/url"
)

// Deploys a Virtual Machine and returns it's id
func (c CloudstackClient) AttachIso(isoid string, vmid string) (AttachIsoResponse, error) {
	var resp AttachIsoResponse
	params := url.Values{}
	params.Set("id", isoid)
	params.Set("virtualmachineid", vmid)

	response, err := NewRequest(c, "attachIso", params)
	if err != nil {
		return resp, err
	}

	resp = response.(AttachIsoResponse)
	return resp, err
}

func (c CloudstackClient) DetachIso(vmid string) (DetachIsoResponse, error) {
	var resp DetachIsoResponse
	params := url.Values{}
	params.Set("virtualmachineid", vmid)
	response, err := NewRequest(c, "detachIso", params)
	if err != nil {
		return resp, err
	}
	resp = response.(DetachIsoResponse)
	return resp, err
}

func (c CloudstackClient) ListIsos() (ListIsosResponse, error) {
	var resp ListIsosResponse
	response, err := NewRequest(c, "listIsos", nil)
	if err != nil {
		return resp, err
	}
	resp = response.(ListIsosResponse)
	return resp, err
}

type DetachIsoResponse struct {
	Detachisoresponse struct {
		Jobid string `json:"jobid"`
	} `json:"detachisoresponse"`
}

type AttachIsoResponse struct {
}

type ListIsosResponse struct {
}
