package gopherstack

import (
	"net/url"
	"strconv"
	"strings"
)

// Add tags to specified resources
func (c CloudstackClient) CreateTags(options *CreateTags) (CreateTagsResponse, error) {
	var resp CreateTagsResponse
	params := url.Values{}

	params.Set("resourceids", strings.Join(options.Resourceids, ","))
	params.Set("resourcetype", options.Resourcetype)
	for j, tag := range options.Tags {
		params.Set("tags["+strconv.Itoa(j+1)+"].key", tag.Key)
		params.Set("tags["+strconv.Itoa(j+1)+"].value", tag.Value)
	}

	if options.Customer != "" {
		params.Set("customer", options.Customer)
	}

	response, err := NewRequest(c, "createTags", params)
	if err != nil {
		return resp, err
	}

	resp = response.(CreateTagsResponse)
	return resp, err
}

// Remove tags from specified resources
func (c CloudstackClient) DeleteTags(options *DeleteTags) (DeleteTagsResponse, error) {
	var resp DeleteTagsResponse
	params := url.Values{}

	params.Set("resourceids", strings.Join(options.Resourceids, ","))
	params.Set("resourcetype", options.Resourcetype)
	for j, tag := range options.Tags {
		params.Set("tags["+strconv.Itoa(j+1)+"].key", tag.Key)
		params.Set("tags["+strconv.Itoa(j+1)+"].value", tag.Value)
	}

	response, err := NewRequest(c, "deleteTags", params)
	if err != nil {
		return resp, err
	}

	resp = response.(DeleteTagsResponse)
	return resp, err
}

// Returns all items with a particular tag
func (c CloudstackClient) ListTags(options *ListTags) (ListTagsResponse, error) {
	var resp ListTagsResponse
	params := url.Values{}
	if options.Account != "" {
		params.Set("account", options.Account)
	}
	if options.Customer != "" {
		params.Set("customer", options.Customer)
	}
	if options.Domainid != "" {
		params.Set("domainid", options.Domainid)
	}
	if options.Isrecursive {
		params.Set("isrecrusive", "true")
	}
	if options.Key != "" {
		params.Set("key", options.Key)
	}
	if options.Keyword != "" {
		params.Set("keyword", options.Keyword)
	}
	if options.Listall {
		params.Set("listall", "true")
	}
	if options.Page != "" {
		params.Set("page", options.Page)
	}
	if options.Pagesize != "" {
		params.Set("pagesize", options.Pagesize)
	}
	if options.Projectid != "" {
		params.Set("projectid", options.Projectid)
	}
	if options.Resourceid != "" {
		params.Set("resourceid", options.Resourceid)
	}
	if options.Resourcetype != "" {
		params.Set("resourcetype", options.Resourcetype)
	}
	if options.Value != "" {
		params.Set("value", options.Value)
	}
	response, err := NewRequest(c, "listTags", params)
	if err != nil {
		return resp, err
	}

	resp = response.(ListTagsResponse)
	return resp, err
}

type Tag struct {
	Account      string `json:"account"`
	Customer     string `json:"customer"`
	Domain       string `json:"domain"`
	Domainid     string `json:"domainid"`
	Key          string `json:"key"`
	Project      string `json:"project"`
	Projectid    string `json:"projectid"`
	Resourceid   string `json:"resourceid"`
	Resourcetype string `json:"resourcetype`
	Value        string `json:"value`
}

type ListTagsResponse struct {
	Listtagsresponse struct {
		Count float64 `json:"count"`
		Tag   []Tag   `json:"tag"`
	} `json:"listtagsresponse"`
}

type ListTags struct {
	Account      string `json:"account"`
	Customer     string `json:"customer"`
	Domainid     string `json:"domainid"`
	Isrecursive  bool   `json:"isrecursive"`
	Key          string `json:"key"`
	Keyword      string `json:"keyword"`
	Listall      bool   `json:"listall"`
	Page         string `json:"page"`
	Pagesize     string `json:"pagesize"`
	Projectid    string `json:"projectid"`
	Resourceid   string `json:"resourceid"`
	Resourcetype string `json:"resourcetype`
	Value        string `json:"value`
}

type CreateTags struct {
	Customer     string   `json:"customer"`
	Resourceids  []string `json:"resourceids"`
	Resourcetype string   `json:"resourcetype`
	Tags         []TagArg `json:"tags`
}

type CreateTagsResponse struct {
	Createtagsresponse struct {
		Displaytext string `json:"displaytext"`
		Success     string `json:"success"`
	} `json:"createtagsresponse"`
}

type DeleteTags struct {
	Resourceids  []string `json:"resourceids"`
	Resourcetype string   `json:"resourcetype`
	Tags         []TagArg `json:"tags`
}

type DeleteTagsResponse struct {
	Deletetagsresponse struct {
		Displaytext string `json:"displaytext"`
		Success     string `json:"success"`
	} `json:"deletetagsresponse"`
}

type TagArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
