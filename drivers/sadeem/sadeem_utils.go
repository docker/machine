package sadeem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL = "http://api.sadeem.io/api/v1/"
	mediaType      = "application/json"
	userAgent      = "docker-machine"
)

// http client to connect sadeem

func NewClient(apiKey, clientId string) *Client {

	httpClient := http.DefaultClient
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		UserAgent: userAgent,
		ApiKey:    apiKey,
		ClientId:  clientId,
	}
	return c
}

func (c *Client) CreateSShKey(name, key string) (string, error) {
	var sshKey SSHKey
	sshKey = SSHKey{Name: name, Key: key}
	body, err := json.Marshal(sshKey)
	url := defaultBaseURL + "sshkeys/"
	resp, err := c.SendRequets(url, "POST", string(body))
	if err != nil {
		return "", err
	} else {
		return ReadCreateResponse(resp)
	}
}

func (c *Client) CreateNewVM(hostName, datacenterId, serviceOfferId, templateId, sSHKey string) (string, error) {
	var newVm VMNew
	url := defaultBaseURL + "vms/"
	userScript := "#!/bin/bash\n\necho \"" + sSHKey + "\" > ~/.ssh/authorized_keys"

	newVm = VMNew{
		Hostname:       hostName,
		DatacenterId:   datacenterId,
		ServiceOfferId: serviceOfferId,
		TemplateId:     templateId,
		PrivateNetwork: false,
		EnableFirewall: false,
		UserScript:     userScript,
	}
	body, _ := json.Marshal(newVm)
	resp, err := c.SendRequets(url, "POST", string(body))

	if err != nil {
		return "", err
	} else {
		return ReadCreateResponse(resp)
	}

}

func (c *Client) GetVmState(vm_id string) (string, error) {
	if vm_id == "" {
		return "", fmt.Errorf("Not valid VM Id")
	}
	url := defaultBaseURL + "vms/" + vm_id

	resp, err := c.SendRequets(url, "GET", "")

	var NewData VMResp
	if err == nil {
		NewData, err = ReadVmResponse(resp)
		if err != nil {
			return "", err
		} else {
			return NewData.Data.Status, nil
		}
	} else {
		return "", err
	}
}

// Vms Actions
func (c *Client) StartVm(vm_id string) error {
	url := defaultBaseURL + "vms/" + vm_id + "/start"
	resp, err := c.SendRequets(url, "POST", "")
	if err != nil {
		return err
	} else {
		ReadActionResponse(resp)
		return nil
	}
}

func (c *Client) StopVm(vm_id string) error {
	url := defaultBaseURL + "vms/" + vm_id + "/shutdown"
	resp, err := c.SendRequets(url, "POST", "")
	if err != nil {
		return err
	} else {
		ReadActionResponse(resp)
		return nil
	}
}

func (c *Client) KillVm(vm_id string) error {
	url := defaultBaseURL + "vms/" + vm_id + "/shutdown?force=true"
	resp, err := c.SendRequets(url, "POST", "")
	if err != nil {
		return err
	} else {
		ReadActionResponse(resp)
		return nil
	}
}

func (c *Client) RebootVm(vm_id string) error {
	url := defaultBaseURL + "vms/" + vm_id + "/reboot"
	resp, err := c.SendRequets(url, "POST", "")
	if err != nil {
		return err
	} else {
		ReadActionResponse(resp)
		return nil
	}
}

func (c *Client) destroyVm(vm_id string) error {
	if vm_id == "" {
		return nil
	}
	url := defaultBaseURL + "vms/" + vm_id
	resp, err := c.SendRequets(url, "DELETE", "")
	if err != nil {
		return err
	} else {
		_, err := ReadActionResponse(resp)
		return err
	}
}

// end of Vms actions

// get Vm IP
func (c *Client) GetVmIP(vm_id string) (string, error) {
	url := defaultBaseURL + "vms/" + vm_id

	resp, err := c.SendRequets(url, "GET", "")

	var NewData VMResp
	if err == nil {
		NewData, err = ReadVmResponse(resp)
		if err != nil {
			return "", err
		} else {
			return NewData.Data.Network[0].IP.String(), nil
		}
	} else {
		return "", err
	}
}

func (c *Client) GetDatacenterId(DcName string) (string, error) {
	url := defaultBaseURL + "datacenters/"
	resp, err := c.SendRequets(url, "GET", "")
	if err != nil {
		return "", err
	}
	Data, err := ReadListResponse(resp)

	if err != nil {
		return "", err
	}

	for i := 0; i < len(Data); i++ {
		if strings.ToLower(Data[i].Name) == strings.ToLower(DcName) {
			return Data[i].ID, nil
		}
	}
	return "", errors.New("datacenter not found")
}

func (c *Client) GetOferrId(DcName, DcId string) (string, error) {
	url := defaultBaseURL + "offers/datacenter/" + DcId
	resp, err := c.SendRequets(url, "GET", "")
	if err != nil {
		return "", err
	}
	Data, err := ReadListResponse(resp)

	if err != nil {
		return "", err
	}

	for i := 0; i < len(Data); i++ {
		if strings.ToLower(Data[i].Name) == strings.ToLower(DcName) {
			return Data[i].ID, nil
		}
	}
	return "", errors.New("Service Offer not found")
}

func (c *Client) GetTemplateId(TemplateName string) (string, error) {
	url := defaultBaseURL + "templates/"
	resp, err := c.SendRequets(url, "GET", "")
	if err != nil {
		return "", err
	}
	Data, err := ReadListResponse(resp)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(Data); i++ {
		if strings.ToLower(Data[i].Name) == strings.ToLower(TemplateName) {
			return Data[i].ID, nil
		}
	}
	return "", errors.New("Template not found")
}

//
// Functions to read response from Sadeem API
//Send Http request
func (c *Client) SendRequets(url, method, body string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	var emptyResp *http.Response

	if err != nil {
		return emptyResp, err
	}
	req.Header.Add("User-Agent", c.UserAgent)
	req.Header.Add("x-api-key", c.ApiKey)
	req.Header.Add("x-client-id", c.ClientId)
	req.Header.Add("content-type", "application/json")

	return c.client.Do(req)
}

// Read response for get vm/$id
func ReadVmResponse(r *http.Response) (VMResp, error) {
	var re1 VMResp
	if r.StatusCode >= 300 {
		return re1, ReadErrorResponse(r)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return re1, err
	}

	err = json.Unmarshal(data, &re1)
	if err != nil {
		return re1, err
	}
	return re1, nil
}

func ReadListResponse(r *http.Response) ([]*ListObject, error) {
	var re1 ListResp
	if r.StatusCode >= 300 {
		return re1.Data, ReadErrorResponse(r)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return re1.Data, err
	}
	err = json.Unmarshal(data, &re1)
	if err != nil {
		return re1.Data, err
	}
	return re1.Data, nil
}

// Read result from cerate, return the Id string
func ReadCreateResponse(r *http.Response) (string, error) {
	var createResponce CreateResponce
	if r.StatusCode >= 300 {
		return "", ReadErrorResponse(r)
	}

	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, &createResponce)
	if err != nil {
		return "", err
	}
	return createResponce.Id, nil
}

//Read Error Response
func ReadErrorResponse(r *http.Response) error {
	var errorMsg ErrorMessage
	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return fmt.Errorf("Undefined error %s", r.Status)
	}
	err = json.Unmarshal(data, &errorMsg)
	if err != nil {
		return fmt.Errorf("Undefined error %s", r.Status)
	}
	return fmt.Errorf(errorMsg.Error.Message)
}

func ReadActionResponse(r *http.Response) (string, error) {
	var createResponce CreateResponce
	if r.StatusCode >= 300 {
		return "", ReadErrorResponse(r)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, &createResponce)
	if err != nil {
		return "", err
	}

	return createResponce.Id, nil
}
