package linodego

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Linode Service
type LinodeService struct {
	client *Client
}

// Response for linode.list API
type LinodesListResponse struct {
	Response
	Linodes []Linode
}

// General Linode API Response
type LinodeResponse struct {
	Response
	LinodeId LinodeId
}

// Job Response
type JobResponse struct {
	Response
	JobId JobId
}

// List all Linodes. If linodeId is less than 0, all linodes are returned.
// Otherwise, only returns the linode for given Id.
func (t *LinodeService) List(linodeId int) (*LinodesListResponse, error) {
	u := &url.Values{}
	v := LinodesListResponse{}
	if linodeId > 0 {
		u.Add("LinodeID", strconv.Itoa(linodeId))
	}
	if err := t.client.do("linode.list", u, &v.Response); err != nil {
		return nil, err
	}

	v.Linodes = make([]Linode, 5)
	if err := json.Unmarshal(v.RawData, &v.Linodes); err != nil {
		return nil, err
	}
	return &v, nil
}

// Create Linode
func (t *LinodeService) Create(dataCenterId int, planId int, paymentTerm int) (*LinodeResponse, error) {
	u := &url.Values{}
	u.Add("DatacenterID", strconv.Itoa(dataCenterId))
	u.Add("PlanID", strconv.Itoa(planId))
	u.Add("PaymentTerm", strconv.Itoa(paymentTerm))
	v := LinodeResponse{}
	if err := t.client.do("linode.create", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.LinodeId); err != nil {
		return nil, err
	}
	return &v, nil
}

// Shutdown Linode
func (t *LinodeService) Shutdown(linodeId int) (*JobResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	v := JobResponse{}
	if err := t.client.do("linode.shutdown", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.JobId); err != nil {
		return nil, err
	}
	return &v, nil
}

// Reboot Linode
func (t *LinodeService) Reboot(linodeId int, configId int) (*JobResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	if configId > 0 {
		u.Add("ConfigID", strconv.Itoa(configId))
	}
	v := JobResponse{}
	if err := t.client.do("linode.reboot", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.JobId); err != nil {
		return nil, err
	}
	return &v, nil
}

// Boot Linode
func (t *LinodeService) Boot(linodeId int, configId int) (*JobResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	if configId > 0 {
		u.Add("ConfigID", strconv.Itoa(configId))
	}
	v := JobResponse{}
	if err := t.client.do("linode.boot", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.JobId); err != nil {
		return nil, err
	}
	return &v, nil
}

// Clone Linode
func (t *LinodeService) Clone(linodeId int, dataCenterId int, planId int, paymentTerm int) (*LinodeResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("DatacenterID", strconv.Itoa(dataCenterId))
	u.Add("PlanID", strconv.Itoa(planId))

	if paymentTerm > 0 {
		u.Add("PaymentTerm", strconv.Itoa(paymentTerm))
	}
	v := LinodeResponse{}
	if err := t.client.do("linode.clone", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.LinodeId); err != nil {
		return nil, err
	}
	return &v, nil
}

// Delete Linode
func (t *LinodeService) Delete(linodeId int, skipChecks bool) (*LinodeResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	if skipChecks {
		u.Add("skipChecks", "1")
	}

	v := LinodeResponse{}
	if err := t.client.do("linode.delete", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.LinodeId); err != nil {
		return nil, err
	}
	return &v, nil
}

// Resize Linode
func (t *LinodeService) Resize(linodeId int, planId int) (*LinodeResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("PlanID", strconv.Itoa(planId))

	v := LinodeResponse{}
	if err := t.client.do("linode.resize", u, &v.Response); err != nil {
		return nil, err
	}

	v.LinodeId.LinodeId = linodeId // hardcode this to input
	return &v, nil
}

// Update Linode
func (t *LinodeService) Update(linodeId int, args map[string]interface{}) (*LinodeResponse, error) {
	u := &url.Values{}

	u.Add("LinodeID", strconv.Itoa(linodeId))
	// add optional parameters
	for k, v := range args {
		u.Add(k, fmt.Sprintf("%v", v))
	}

	v := LinodeResponse{}
	if err := t.client.do("linode.update", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.LinodeId); err != nil {
		return nil, err
	}
	return &v, nil
}
