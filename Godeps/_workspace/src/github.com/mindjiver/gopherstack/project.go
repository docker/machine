package gopherstack

import (
	"net/url"
)

// List the available Cloudstack projects
func (c CloudstackClient) ListProjects(name string) (ListProjectsResponse, error) {
	var resp ListProjectsResponse
	params := url.Values{}

	if name != "" {
		params.Set("name", name)
	}
	response, err := NewRequest(c, "listProjects", params)
	if err != nil {
		return resp, err
	}

	resp = response.(ListProjectsResponse)
	return resp, err
}

type Project struct {
	Account     string        `json:"account"`
	Displaytext string        `json:"displaytext"`
	Domain      string        `json:"domain"`
	Domainid    string        `json:"domainid"`
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	State       string        `json:"state"`
	Tags        []interface{} `json:"tags"`
}

type ListProjectsResponse struct {
	Listprojectsresponse struct {
		Count   float64   `json:"count"`
		Project []Project `json:"project"`
	} `json:"listprojectsresponse"`
}
