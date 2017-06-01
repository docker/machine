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
)

// Contains functionality for deployments API.
type DeploymentsAPI struct {
	client *Client
}

var deploymentUrl string = "/deployments"

// Creates a deployment
func (api *DeploymentsAPI) Create(deploymentSpec *DeploymentCreateSpec) (task *Task, err error) {
	body, err := json.Marshal(deploymentSpec)
	if err != nil {
		return
	}
	res, err := api.client.restClient.Post(
		api.client.Endpoint+deploymentUrl,
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

// Deletes a deployment with specified ID.
func (api *DeploymentsAPI) Delete(id string) (task *Task, err error) {
	res, err := api.client.restClient.Delete(api.getEntityUrl(id), api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

// Deploys a deployment with specified ID.
func (api *DeploymentsAPI) Deploy(id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/deploy",
		"application/json",
		bytes.NewBuffer([]byte("")),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

// Destroys a deployment with specified ID.
func (api *DeploymentsAPI) Destroy(id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/destroy",
		"application/json",
		bytes.NewBuffer([]byte("")),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

// Returns all deployments.
func (api *DeploymentsAPI) GetAll() (result *Deployments, err error) {
	res, err := api.client.restClient.Get(api.client.Endpoint+deploymentUrl, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	result = &Deployments{}
	err = json.NewDecoder(res.Body).Decode(result)
	return
}

// Gets a deployment with the specified ID.
func (api *DeploymentsAPI) Get(id string) (deployment *Deployment, err error) {
	res, err := api.client.restClient.Get(api.getEntityUrl(id), api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	res, err = getError(res)
	if err != nil {
		return
	}
	var result Deployment
	err = json.NewDecoder(res.Body).Decode(&result)
	return &result, nil
}

// Gets all hosts with the specified deployment ID.
func (api *DeploymentsAPI) GetHosts(id string) (result *Hosts, err error) {
	uri := api.getEntityUrl(id) + "/hosts"
	res, err := api.client.restClient.GetList(api.client.Endpoint, uri, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	result = &Hosts{}
	err = json.Unmarshal(res, result)
	return
}

// Gets all the vms with the specified deployment ID.
func (api *DeploymentsAPI) GetVms(id string) (result *VMs, err error) {
	uri := api.getEntityUrl(id) + "/vms"
	res, err := api.client.restClient.GetList(api.client.Endpoint, uri, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	result = &VMs{}
	err = json.Unmarshal(res, result)
	return
}

// Initialize deployment migration from source to destination
func (api *DeploymentsAPI) InitializeDeploymentMigration(sourceAddress string, id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/initialize_migration",
		"application/json",
		bytes.NewBuffer([]byte(sourceAddress)),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

// Finalize deployment migration from source to destination
func (api *DeploymentsAPI) FinalizeDeploymentMigration(sourceAddress string, id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/finalize_migration",
		"application/json",
		bytes.NewBuffer([]byte(sourceAddress)),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()
	task, err = getTask(getError(res))
	return
}

func (api *DeploymentsAPI) SetSecurityGroups(id string, securityGroups *SecurityGroupsSpec) (*Task, error) {
	return setSecurityGroups(api.client, api.getEntityUrl(id), securityGroups)
}

// Update image datastores of a deployment.
func (api *DeploymentsAPI) SetImageDatastores(id string, imageDatastores *ImageDatastores) (task *Task, err error) {
	body, err := json.Marshal(imageDatastores)
	if err != nil {
		return
	}

	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/set_image_datastores",
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

// Pause system with specified deployment ID.
func (api *DeploymentsAPI) PauseSystem(id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/pause_system",
		"application/json",
		bytes.NewBuffer([]byte("")),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()

	task, err = getTask(getError(res))
	return
}

// Pause background tasks of system with specified deployment ID.
func (api *DeploymentsAPI) PauseBackgroundTasks(id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/pause_background_tasks",
		"application/json",
		bytes.NewBuffer([]byte("")),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()

	task, err = getTask(getError(res))
	return
}

// Pause background tasks of system with specified deployment ID.
func (api *DeploymentsAPI) ResumeSystem(id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/resume_system",
		"application/json",
		bytes.NewBuffer([]byte("")),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	defer res.Body.Close()

	task, err = getTask(getError(res))
	return
}

//  Enable cluster type with specified deployment ID.
func (api *DeploymentsAPI) EnableClusterType(id string, clusterConfigSpec *ClusterConfigurationSpec) (result *ClusterConfiguration, err error) {
	body, err := json.Marshal(clusterConfigSpec)
	if err != nil {
		return
	}
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/enable_cluster_type",
		"application/json",
		bytes.NewBuffer(body),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}
	res, err = getError(res)
	if err != nil {
		return
	}
	result = &ClusterConfiguration{}
	err = json.NewDecoder(res.Body).Decode(result)
	return
}

//  Disable cluster type with specified deployment ID.
func (api *DeploymentsAPI) DisableClusterType(id string, clusterConfigSpec *ClusterConfigurationSpec) (task *Task, err error) {
	body, err := json.Marshal(clusterConfigSpec)
	if err != nil {
		return
	}
	res, err := api.client.restClient.Post(
		api.getEntityUrl(id)+"/disable_cluster_type",
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

func (api *DeploymentsAPI) getEntityUrl(id string) (url string) {
	return api.client.Endpoint + deploymentUrl + "/" + id
}
