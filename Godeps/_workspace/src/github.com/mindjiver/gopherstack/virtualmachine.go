package gopherstack

import (
	"encoding/base64"
	"net/url"
	"strings"
)

// Deploys a Virtual Machine and returns it's id
func (c CloudstackClient) DeployVirtualMachine(serviceofferingid string, templateid string, zoneid string, account string, diskofferingid string, displayname string, networkids []string, keypair string, projectid string, userdata string, hypervisor string) (DeployVirtualMachineResponse, error) {
	var resp DeployVirtualMachineResponse

	params := url.Values{}
	params.Set("serviceofferingid", serviceofferingid)
	params.Set("templateid", templateid)
	params.Set("zoneid", zoneid)
	//	params.Set("account", account)
	params.Set("diskofferingid", diskofferingid)
	params.Set("displayname", displayname)
	params.Set("hypervisor", hypervisor)
	if len(networkids) > 0 {
		params.Set("networkids", strings.Join(networkids, ","))
	}
	params.Set("keypair", keypair)
	//	params.Set("projectid", projectid)
	if userdata != "" {
		params.Set("userdata", base64.StdEncoding.EncodeToString([]byte(userdata)))
	}
	response, err := NewRequest(c, "deployVirtualMachine", params)
	if err != nil {
		return resp, err
	}
	resp = response.(DeployVirtualMachineResponse)
	return resp, nil
}

func (c CloudstackClient) UpdateVirtualMachine(id string, displayname string, group string, haenable string, ostypeid string, userdata string) (UpdateVirtualMachineResponse, error) {
	var resp UpdateVirtualMachineResponse

	params := url.Values{}
	params.Set("id", id)
	params.Set("displayname", displayname)
	//	params.Set("group", string)
	//	params.Set("haenable", haenable)
	//	params.Set("ostypeid", ostypeid)
	params.Set("userdata", base64.StdEncoding.EncodeToString([]byte(userdata)))
	response, err := NewRequest(c, "updateVirtualMachine", params)
	if err != nil {
		return resp, err
	}
	resp = response.(UpdateVirtualMachineResponse)
	return resp, err
}

// Stops a Virtual Machine
func (c CloudstackClient) StopVirtualMachine(id string) (StopVirtualMachineResponse, error) {
	var resp StopVirtualMachineResponse
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "stopVirtualMachine", params)
	if err != nil {
		return resp, err
	}
	resp = response.(StopVirtualMachineResponse)
	return resp, err
}

// Destroys a Virtual Machine
func (c CloudstackClient) DestroyVirtualMachine(id string) (DestroyVirtualMachineResponse, error) {
	var resp DestroyVirtualMachineResponse
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "destroyVirtualMachine", params)
	if err != nil {
		return resp, err
	}
	resp = response.(DestroyVirtualMachineResponse)
	return resp, nil
}

// Returns Cloudstack string representation of the Virtual Machine state
func (c CloudstackClient) ListVirtualMachines(id string) (ListVirtualMachinesResponse, error) {
	var resp ListVirtualMachinesResponse
	params := url.Values{}
	params.Set("id", id)
	response, err := NewRequest(c, "listVirtualMachines", params)
	if err != nil {
		return resp, err
	}
	resp = response.(ListVirtualMachinesResponse)
	return resp, err
}

type DeployVirtualMachineResponse struct {
	Deployvirtualmachineresponse struct {
		ID    string `json:"id"`
		Jobid string `json:"jobid"`
	} `json:"deployvirtualmachineresponse"`
}

type DestroyVirtualMachineResponse struct {
	Destroyvirtualmachineresponse struct {
		Jobid string `json:"jobid"`
	} `json:"destroyvirtualmachineresponse"`
}

type StopVirtualMachineResponse struct {
	Stopvirtualmachineresponse struct {
		Jobid string `json:"jobid"`
	} `json:"stopvirtualmachineresponse"`
}

type Nic struct {
	Gateway     string `json:"gateway"`
	ID          string `json:"id"`
	Ipaddress   string `json:"ipaddress"`
	Isdefault   bool   `json:"isdefault"`
	Macaddress  string `json:"macaddress"`
	Netmask     string `json:"netmask"`
	Networkid   string `json:"networkid"`
	Traffictype string `json:"traffictype"`
	Type        string `json:"type"`
}

type Virtualmachine struct {
	Account             string        `json:"account"`
	Cpunumber           float64       `json:"cpunumber"`
	Cpuspeed            float64       `json:"cpuspeed"`
	Created             string        `json:"created"`
	Displayname         string        `json:"displayname"`
	Domain              string        `json:"domain"`
	Domainid            string        `json:"domainid"`
	Guestosid           string        `json:"guestosid"`
	Haenable            bool          `json:"haenable"`
	Hypervisor          string        `json:"hypervisor"`
	ID                  string        `json:"id"`
	IsoId               string        `json:"isoid"`
	IsoName             string        `json:"isoname"`
	Keypair             string        `json:"keypair"`
	Memory              float64       `json:"memory"`
	Name                string        `json:"name"`
	Nic                 []Nic         `json:"nic"`
	Passwordenabled     bool          `json:"passwordenabled"`
	Rootdeviceid        float64       `json:"rootdeviceid"`
	Rootdevicetype      string        `json:"rootdevicetype"`
	Securitygroup       []interface{} `json:"securitygroup"`
	Serviceofferingid   string        `json:"serviceofferingid"`
	Serviceofferingname string        `json:"serviceofferingname"`
	State               string        `json:"state"`
	Tags                []interface{} `json:"tags"`
	Templatedisplaytext string        `json:"templatedisplaytext"`
	Templateid          string        `json:"templateid"`
	Templatename        string        `json:"templatename"`
	Zoneid              string        `json:"zoneid"`
	Zonename            string        `json:"zonename"`
}

type ListVirtualMachinesResponse struct {
	Listvirtualmachinesresponse struct {
		Count          float64          `json:"count"`
		Virtualmachine []Virtualmachine `json:"virtualmachine"`
	} `json:"listvirtualmachinesresponse"`
}

type UpdateVirtualMachineResponse struct {
	Updatevirtualmachineresponse struct {
	}
}
