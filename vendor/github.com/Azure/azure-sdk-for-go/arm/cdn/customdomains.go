package cdn

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator 0.14.0.0
// Changes may cause incorrect behavior and will be lost if the code is
// regenerated.

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"net/http"
	"net/url"
)

// CustomDomainsClient is the use these APIs to manage Azure CDN resources
// through the Azure Resource Manager. You must make sure that requests made
// to these resources are secure. For more information, see <a
// href="https://msdn.microsoft.com/en-us/library/azure/dn790557.aspx">Authenticating
// Azure Resource Manager requests.</a>
type CustomDomainsClient struct {
	ManagementClient
}

// NewCustomDomainsClient creates an instance of the CustomDomainsClient
// client.
func NewCustomDomainsClient(subscriptionID string) CustomDomainsClient {
	return NewCustomDomainsClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewCustomDomainsClientWithBaseURI creates an instance of the
// CustomDomainsClient client.
func NewCustomDomainsClientWithBaseURI(baseURI string, subscriptionID string) CustomDomainsClient {
	return CustomDomainsClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// Create sends the create request.
//
// customDomainName is name of the custom domain within an endpoint
// customDomainProperties is custom domain properties required for creation
// endpointName is name of the endpoint within the CDN profile profileName is
// name of the CDN profile within the resource group resourceGroupName is
// name of the resource group within the Azure subscription
func (client CustomDomainsClient) Create(customDomainName string, customDomainProperties CustomDomainParameters, endpointName string, profileName string, resourceGroupName string) (result CustomDomain, ae error) {
	req, err := client.CreatePreparer(customDomainName, customDomainProperties, endpointName, profileName, resourceGroupName)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Create", nil, "Failure preparing request")
	}

	resp, err := client.CreateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Create", resp, "Failure sending request")
	}

	result, err = client.CreateResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Create", resp, "Failure responding to request")
	}

	return
}

// CreatePreparer prepares the Create request.
func (client CustomDomainsClient) CreatePreparer(customDomainName string, customDomainProperties CustomDomainParameters, endpointName string, profileName string, resourceGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"customDomainName":  url.QueryEscape(customDomainName),
		"endpointName":      url.QueryEscape(endpointName),
		"profileName":       url.QueryEscape(profileName),
		"resourceGroupName": url.QueryEscape(resourceGroupName),
		"subscriptionId":    url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPut(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Cdn/profiles/{profileName}/endpoints/{endpointName}/customDomains/{customDomainName}"),
		autorest.WithJSON(customDomainProperties),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// CreateSender sends the Create request. The method will close the
// http.Response Body if it receives an error.
func (client CustomDomainsClient) CreateSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// CreateResponder handles the response to the Create request. The method always
// closes the http.Response Body.
func (client CustomDomainsClient) CreateResponder(resp *http.Response) (result CustomDomain, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated, http.StatusAccepted),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// DeleteIfExists sends the delete if exists request.
//
// customDomainName is name of the custom domain within an endpoint
// endpointName is name of the endpoint within the CDN profile profileName is
// name of the CDN profile within the resource group resourceGroupName is
// name of the resource group within the Azure subscription
func (client CustomDomainsClient) DeleteIfExists(customDomainName string, endpointName string, profileName string, resourceGroupName string) (result CustomDomain, ae error) {
	req, err := client.DeleteIfExistsPreparer(customDomainName, endpointName, profileName, resourceGroupName)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "DeleteIfExists", nil, "Failure preparing request")
	}

	resp, err := client.DeleteIfExistsSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "DeleteIfExists", resp, "Failure sending request")
	}

	result, err = client.DeleteIfExistsResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "DeleteIfExists", resp, "Failure responding to request")
	}

	return
}

// DeleteIfExistsPreparer prepares the DeleteIfExists request.
func (client CustomDomainsClient) DeleteIfExistsPreparer(customDomainName string, endpointName string, profileName string, resourceGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"customDomainName":  url.QueryEscape(customDomainName),
		"endpointName":      url.QueryEscape(endpointName),
		"profileName":       url.QueryEscape(profileName),
		"resourceGroupName": url.QueryEscape(resourceGroupName),
		"subscriptionId":    url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsDelete(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Cdn/profiles/{profileName}/endpoints/{endpointName}/customDomains/{customDomainName}"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// DeleteIfExistsSender sends the DeleteIfExists request. The method will close the
// http.Response Body if it receives an error.
func (client CustomDomainsClient) DeleteIfExistsSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// DeleteIfExistsResponder handles the response to the DeleteIfExists request. The method always
// closes the http.Response Body.
func (client CustomDomainsClient) DeleteIfExistsResponder(resp *http.Response) (result CustomDomain, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusAccepted, http.StatusNoContent),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Get sends the get request.
//
// customDomainName is name of the custom domain within an endpoint
// endpointName is name of the endpoint within the CDN profile profileName is
// name of the CDN profile within the resource group resourceGroupName is
// name of the resource group within the Azure subscription
func (client CustomDomainsClient) Get(customDomainName string, endpointName string, profileName string, resourceGroupName string) (result CustomDomain, ae error) {
	req, err := client.GetPreparer(customDomainName, endpointName, profileName, resourceGroupName)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Get", nil, "Failure preparing request")
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Get", resp, "Failure sending request")
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client CustomDomainsClient) GetPreparer(customDomainName string, endpointName string, profileName string, resourceGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"customDomainName":  url.QueryEscape(customDomainName),
		"endpointName":      url.QueryEscape(endpointName),
		"profileName":       url.QueryEscape(profileName),
		"resourceGroupName": url.QueryEscape(resourceGroupName),
		"subscriptionId":    url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Cdn/profiles/{profileName}/endpoints/{endpointName}/customDomains/{customDomainName}"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client CustomDomainsClient) GetSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client CustomDomainsClient) GetResponder(resp *http.Response) (result CustomDomain, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListByEndpoint sends the list by endpoint request.
//
// endpointName is name of the endpoint within the CDN profile profileName is
// name of the CDN profile within the resource group resourceGroupName is
// name of the resource group within the Azure subscription
func (client CustomDomainsClient) ListByEndpoint(endpointName string, profileName string, resourceGroupName string) (result CustomDomainListResult, ae error) {
	req, err := client.ListByEndpointPreparer(endpointName, profileName, resourceGroupName)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "ListByEndpoint", nil, "Failure preparing request")
	}

	resp, err := client.ListByEndpointSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "ListByEndpoint", resp, "Failure sending request")
	}

	result, err = client.ListByEndpointResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "ListByEndpoint", resp, "Failure responding to request")
	}

	return
}

// ListByEndpointPreparer prepares the ListByEndpoint request.
func (client CustomDomainsClient) ListByEndpointPreparer(endpointName string, profileName string, resourceGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"endpointName":      url.QueryEscape(endpointName),
		"profileName":       url.QueryEscape(profileName),
		"resourceGroupName": url.QueryEscape(resourceGroupName),
		"subscriptionId":    url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Cdn/profiles/{profileName}/endpoints/{endpointName}/customDomains"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// ListByEndpointSender sends the ListByEndpoint request. The method will close the
// http.Response Body if it receives an error.
func (client CustomDomainsClient) ListByEndpointSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// ListByEndpointResponder handles the response to the ListByEndpoint request. The method always
// closes the http.Response Body.
func (client CustomDomainsClient) ListByEndpointResponder(resp *http.Response) (result CustomDomainListResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Update sends the update request.
//
// customDomainName is name of the custom domain within an endpoint
// customDomainProperties is custom domain properties to update endpointName
// is name of the endpoint within the CDN profile profileName is name of the
// CDN profile within the resource group resourceGroupName is name of the
// resource group within the Azure subscription
func (client CustomDomainsClient) Update(customDomainName string, customDomainProperties CustomDomainParameters, endpointName string, profileName string, resourceGroupName string) (result ErrorResponse, ae error) {
	req, err := client.UpdatePreparer(customDomainName, customDomainProperties, endpointName, profileName, resourceGroupName)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Update", nil, "Failure preparing request")
	}

	resp, err := client.UpdateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Update", resp, "Failure sending request")
	}

	result, err = client.UpdateResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "cdn/CustomDomainsClient", "Update", resp, "Failure responding to request")
	}

	return
}

// UpdatePreparer prepares the Update request.
func (client CustomDomainsClient) UpdatePreparer(customDomainName string, customDomainProperties CustomDomainParameters, endpointName string, profileName string, resourceGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"customDomainName":  url.QueryEscape(customDomainName),
		"endpointName":      url.QueryEscape(endpointName),
		"profileName":       url.QueryEscape(profileName),
		"resourceGroupName": url.QueryEscape(resourceGroupName),
		"subscriptionId":    url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPatch(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Cdn/profiles/{profileName}/endpoints/{endpointName}/customDomains/{customDomainName}"),
		autorest.WithJSON(customDomainProperties),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// UpdateSender sends the Update request. The method will close the
// http.Response Body if it receives an error.
func (client CustomDomainsClient) UpdateSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// UpdateResponder handles the response to the Update request. The method always
// closes the http.Response Body.
func (client CustomDomainsClient) UpdateResponder(resp *http.Response) (result ErrorResponse, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
