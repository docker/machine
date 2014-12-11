package gopherstack

import (
	"net/url"
)

func (c CloudstackClient) ListDiskOfferings(domainid string, id string, keyword string, name string, page string, pagesize string) (ListDiskOfferingsResponse, error) {
	var resp ListDiskOfferingsResponse
	params := url.Values{}
	//params.Set("domainid", domainid)
	response, err := NewRequest(c, "listDiskOfferings", params)
	if err != nil {
		return resp, err
	}
	resp = response.(ListDiskOfferingsResponse)
	return resp, err
}

type DiskOffering struct {
	Created      string  `json:"created"`
	Disksize     float64 `json:"disksize"`
	Displaytext  string  `json:"displaytext"`
	ID           string  `json:"id"`
	Iscustomized bool    `json:"iscustomized"`
	Name         string  `json:"name"`
	Storagetype  string  `json:"storagetype"`
}

type ListDiskOfferingsResponse struct {
	Listdiskofferingsresponse struct {
		Count        float64        `json:"count"`
		Diskoffering []DiskOffering `json:"diskoffering"`
	} `json:"listdiskofferingsresponse"`
}
