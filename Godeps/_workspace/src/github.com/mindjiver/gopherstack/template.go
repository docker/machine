package gopherstack

import (
	"net/url"
)

// Creates a Template of a Virtual Machine by it's ID
func (c CloudstackClient) CreateTemplate(options *CreateTemplate) (CreateTemplateResponse, error) {
	var resp CreateTemplateResponse
	params := url.Values{}
	params.Set("displaytext", options.Displaytext)
	params.Set("name", options.Name)
	params.Set("ostypeid", options.Ostypeid)
	if options.Volumeid != "" {
		params.Set("volumeid", options.Volumeid)
	}
	if options.Snapshotid != "" {
		params.Set("snapshotid", options.Snapshotid)
	}

	if options.Isdynamicallyscalable {
		params.Set("isdynamicallyscalable", "true")
	}
	if options.Isextractable {
		params.Set("isextractable", "true")
	}
	if options.Isfeatured {
		params.Set("isfeatured", "true")
	}
	if options.Ispublic {
		params.Set("ispublic", "true")
	}
	if options.Passwordenabled {
		params.Set("passwordenabled", "true")
	}

	response, err := NewRequest(c, "createTemplate", params)
	if err != nil {
		return resp, err
	}

	resp = response.(CreateTemplateResponse)
	return resp, err
}

// Returns all available templates
func (c CloudstackClient) ListTemplates(name string, filter string) (ListTemplatesResponse, error) {
	var resp ListTemplatesResponse
	params := url.Values{}
	params.Set("name", name)
	params.Set("templatefilter", filter)
	response, err := NewRequest(c, "listTemplates", params)
	if err != nil {
		return resp, err
	}

	resp = response.(ListTemplatesResponse)
	return resp, err
}

// Deletes an template by its ID.
func (c CloudstackClient) DeleteTemplate(id string) (DeleteTemplateResponse, error) {
	var resp DeleteTemplateResponse
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "deleteTemplate", params)
	if err != nil {
		return resp, err
	}

	resp = response.(DeleteTemplateResponse)
	return resp, err
}

type CreateTemplateResponse struct {
	Createtemplateresponse struct {
		ID    string `json:"id"`
		Jobid string `json:"jobid"`
	} `json:"createtemplateresponse"`
}

type DeleteTemplateResponse struct {
	Deletetemplateresponse struct {
	}
}

type Template struct {
	Account          string  `json:"account"`
	Created          string  `json:"created"`
	CrossZones       bool    `json:"crossZones"`
	Displaytext      string  `json:"displaytext"`
	Domain           string  `json:"domain"`
	Domainid         string  `json:"domainid"`
	Format           string  `json:"format"`
	Hypervisor       string  `json:"hypervisor"`
	ID               string  `json:"id"`
	Isextractable    bool    `json:"isextractable"`
	Isfeatured       bool    `json:"isfeatured"`
	Ispublic         bool    `json:"ispublic"`
	Isready          bool    `json:"isready"`
	Name             string  `json:"name"`
	Ostypeid         string  `json:"ostypeid"`
	Ostypename       string  `json:"ostypename"`
	Passwordenabled  bool    `json:"passwordenabled"`
	Size             float64 `json:"size"`
	Sourcetemplateid string  `json:"sourcetemplateid"`
	Sshkeyenabled    bool    `json:"sshkeyenabled"`
	Status           string  `json:"status"`
	Tags             []Tag   `json:"tags"`
	Templatetype     string  `json:"templatetype"`
	Zoneid           string  `json:"zoneid"`
	Zonename         string  `json:"zonename"`
}

type ListTemplatesResponse struct {
	Listtemplatesresponse struct {
		Count    float64    `json:"count"`
		Template []Template `json:"template"`
	} `json:"listtemplatesresponse"`
}

type CreateTemplate struct {
	Displaytext           string `json:"displaytext"`
	Isdynamicallyscalable bool   `json:"isdynamicallyscalable"`
	Isextractable         bool   `json:"isextractable"`
	Isfeatured            bool   `json:"isfeatured"`
	Ispublic              bool   `json:"ispublic"`
	Name                  string `json:"name"`
	Ostypeid              string `json:"ostypeid"`
	Passwordenabled       bool   `json:"passwordenabled"`
	Snapshotid            string `json:"snapshotid"`
	Volumeid              string `json:"volumeid"`
}
