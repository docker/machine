// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package photon

import (
	"bytes"
	"encoding/json"

	"github.com/vmware/photon-controller-go-sdk/photon/internal/rest"
)

// Contains functionality for hosts API.
type HostsAPI struct {
	client *Client
}

var hostUrl string = "/hosts"

// Creates a host.
func (api *HostsAPI) Create(hostSpec *HostCreateSpec, deploymentId string) (task *Task, err error) {
	body, err := json.Marshal(hostSpec)
	if err != nil {
		return
	}
	res, err := rest.Post(api.client.httpClient,
		api.client.Endpoint+deploymentUrl+"/"+deploymentId+hostUrl,
		"application/json",
		bytes.NewBuffer(body),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

// Deletes a host with specified ID.
func (api *HostsAPI) Delete(id string) (task *Task, err error) {
	res, err := rest.Delete(api.client.httpClient, api.client.Endpoint+hostUrl+"/"+id, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

// Returns all hosts
func (api *HostsAPI) GetAll() (result *Hosts, err error) {
	res, err := rest.Get(api.client.httpClient, api.client.Endpoint+hostUrl, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	result = &Hosts{}
	err = json.NewDecoder(res.Body).Decode(result)
	return
}

// Gets a host with the specified ID.
func (api *HostsAPI) Get(id string) (host *Host, err error) {
	res, err := rest.Get(api.client.httpClient, api.client.Endpoint+hostUrl+"/"+id, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	var result Host
	err = json.NewDecoder(res.Body).Decode(&result)
	return &result, nil
}

// Sets host's availability zone.
func (api *HostsAPI) SetAvailabilityZone(id string, availabilityZone *HostSetAvailabilityZoneOperation) (task *Task, err error) {
	body, err := json.Marshal(availabilityZone)
	if err != nil {
		return
	}

	res, err := rest.Post(api.client.httpClient,
		api.client.Endpoint+hostUrl+"/"+id+"/set_availability_zone",
		"application/json",
		bytes.NewBuffer(body),
		api.client.options.TokenOptions.AccessToken)

	if err != nil {
		return
	}

	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

// Gets all tasks with the specified host ID, using options to filter the results.
// If options is nil, no filtering will occur.
func (api *HostsAPI) GetTasks(id string, options *TaskGetOptions) (result *TaskList, err error) {
	uri := api.client.Endpoint + hostUrl + "/" + id + "/tasks"
	if options != nil {
		uri += getQueryString(options)
	}
	res, err := rest.GetList(api.client.httpClient, api.client.Endpoint, uri, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	result = &TaskList{}
	err = json.Unmarshal(res, result)
	return
}

// Gets all the vms with the specified deployment ID.
func (api *HostsAPI) GetVMs(id string) (result *VMs, err error) {
	res, err := rest.Get(api.client.httpClient, api.client.Endpoint+hostUrl+"/"+id+"/vms", api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	result = &VMs{}
	err = json.NewDecoder(res.Body).Decode(result)
	return
}
