package egoscale

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func (exo *Client) CreateVirtualMachine(p MachineProfile) (string, error) {

	params := url.Values{}
	params.Set("serviceofferingid", p.ServiceOffering)
	params.Set("templateid", p.Template)
	params.Set("zoneid", p.Zone)

	params.Set("displayname", p.Name)
	if len(p.Userdata) > 0 {
		params.Set("userdata", base64.StdEncoding.EncodeToString([]byte(p.Userdata)))
	}
	if len(p.Keypair) > 0 {
		params.Set("keypair", p.Keypair)
	}

	params.Set("securitygroupids", strings.Join(p.SecurityGroups, ","))

	resp, err := exo.Request("deployVirtualMachine", params)

	if err != nil {
		return "", err
	}

	var r DeployVirtualMachineResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobID, nil
}

func (exo *Client) StartVirtualMachine(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)

	resp, err := exo.Request("startVirtualMachine", params)

	if err != nil {
		return "", err
	}

	var r StartVirtualMachineResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobID, nil
}

func (exo *Client) StopVirtualMachine(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)

	resp, err := exo.Request("stopVirtualMachine", params)

	if err != nil {
		return "", err
	}

	var r StopVirtualMachineResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobID, nil
}

func (exo *Client) RebootVirtualMachine(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)

	resp, err := exo.Request("rebootVirtualMachine", params)

	if err != nil {
		return "", err
	}

	var r RebootVirtualMachineResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobID, nil
}

func (exo *Client) DestroyVirtualMachine(id string) (string, error) {
	params := url.Values{}
	params.Set("id", id)

	resp, err := exo.Request("destroyVirtualMachine", params)

	if err != nil {
		return "", err
	}

	var r DestroyVirtualMachineResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}

	return r.JobID, nil
}

func (exo *Client) GetVirtualMachine(id string) (*VirtualMachine, error) {

	params := url.Values{}
	params.Set("id", id)

	resp, err := exo.Request("listVirtualMachines", params)

	if err != nil {
		return nil, err
	}

	var r ListVirtualMachinesResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	if len(r.VirtualMachines) == 1 {
		machine := r.VirtualMachines[0]
		return machine, nil
	} else {
		return nil, fmt.Errorf("cannot retrieve virtualmachine with id %s", id)
	}
}

func (exo *Client) ListVirtualMachines(id string) ([]*VirtualMachine, error) {

	params := url.Values{}
	params.Set("id", id)

	resp, err := exo.Request("listVirtualMachines", params)

	if err != nil {
		return nil, err
	}

	var r ListVirtualMachinesResponse

	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return r.VirtualMachines, nil
}
