package linodego

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// Linode IP Service
type LinodeIPService struct {
	client *Client
}

// IP List Response
type LinodeIPListResponse struct {
	Response
	FullIPAddresses []FullIPAddress
}

// IP Address Response
type LinodeIPAddressResponse struct {
	Response
	IPAddress IPAddress
}

// IP Address with RDNS Response
type LinodeRDNSIPAddressResponse struct {
	Response
	RDNSIPAddress RDNSIPAddress
}

// Full IP Address Response
type LinodeLinodeIPAddressResponse struct {
	Response
	LinodeIPAddresses []LinodeIPAddress
}

// List All Ips. If linodeId or ipAddressId is less than 0, all IPs are returned.
// Otherwise, limits the reuslt to the given linodeId, ipAddressId or both.
func (t *LinodeIPService) List(linodeId int, ipAddressId int) (*LinodeIPListResponse, error) {
	u := &url.Values{}
	v := LinodeIPListResponse{}
	if linodeId > 0 {
		u.Add("LinodeID", strconv.Itoa(linodeId))
	}
	if ipAddressId > 0 {
		u.Add("IPAddressID", strconv.Itoa(ipAddressId))
	}
	if err := t.client.do("linode.ip.list", u, &v.Response); err != nil {
		return nil, err
	}

	v.FullIPAddresses = make([]FullIPAddress, 5)
	if err := json.Unmarshal(v.RawData, &v.FullIPAddresses); err != nil {
		return nil, err
	}
	return &v, nil
}

// Add Private IP
func (t *LinodeIPService) AddPrivate(linodeId int) (*LinodeIPAddressResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	v := LinodeIPAddressResponse{}
	if err := t.client.do("linode.ip.addprivate", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.IPAddress); err != nil {
		return nil, err
	}
	return &v, nil
}

// Add Public IP
func (t *LinodeIPService) AddPublic(linodeId int) (*LinodeIPAddressResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	v := LinodeIPAddressResponse{}
	if err := t.client.do("linode.ip.addpublic", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.IPAddress); err != nil {
		return nil, err
	}
	return &v, nil
}

// Set RDNS
func (t *LinodeIPService) SetRDNS(linodeId int, hostname string) (*LinodeRDNSIPAddressResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("Hostname", hostname)
	v := LinodeRDNSIPAddressResponse{}
	if err := t.client.do("linode.ip.setrdns", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.RDNSIPAddress); err != nil {
		return nil, err
	}
	return &v, nil
}

// Swap Ips
func (t *LinodeIPService) Swap(ipAddressId int, withIPAddressId int, toLinodeId int) (*LinodeLinodeIPAddressResponse, error) {
	u := &url.Values{}
	u.Add("toLinodeID", strconv.Itoa(toLinodeId))
	u.Add("ipAddressID", strconv.Itoa(ipAddressId))
	u.Add("withIPAddressID", strconv.Itoa(withIPAddressId))
	v := LinodeLinodeIPAddressResponse{}
	if err := t.client.do("linode.ip.swap", u, &v.Response); err != nil {
		return nil, err
	}

	v.LinodeIPAddresses = make([]LinodeIPAddress, 2)
	if err := json.Unmarshal(v.RawData, &v.LinodeIPAddresses); err != nil {
		return nil, err
	}
	return &v, nil
}
