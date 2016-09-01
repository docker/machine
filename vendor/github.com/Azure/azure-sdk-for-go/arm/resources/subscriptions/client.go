// Package subscriptions implements the Azure ARM Subscriptions service API
// version 2015-11-01.
//
package subscriptions

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

const (
	// APIVersion is the version of the Subscriptions
	APIVersion = "2015-11-01"

	// DefaultBaseURI is the default URI used for the service Subscriptions
	DefaultBaseURI = "https://management.azure.com"
)

// ManagementClient is the base client for Subscriptions.
type ManagementClient struct {
	autorest.Client
	BaseURI        string
	SubscriptionID string
}

// New creates an instance of the ManagementClient client.
func New(subscriptionID string) ManagementClient {
	return NewWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewWithBaseURI creates an instance of the ManagementClient client.
func NewWithBaseURI(baseURI string, subscriptionID string) ManagementClient {
	return ManagementClient{
		Client:         autorest.NewClientWithUserAgent(UserAgent()),
		BaseURI:        baseURI,
		SubscriptionID: subscriptionID,
	}
}

// Get gets details about particular subscription.
//
// subscriptionID is id of the subscription.
func (client ManagementClient) Get(subscriptionID string) (result Subscription, err error) {
	req, err := client.GetPreparer(subscriptionID)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "Get", nil, "Failure preparing request")
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "Get", resp, "Failure sending request")
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client ManagementClient) GetPreparer(subscriptionID string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(subscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client ManagementClient) GetSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client ManagementClient) GetResponder(resp *http.Response) (result Subscription, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// List gets a list of the subscriptionIds.
func (client ManagementClient) List() (result SubscriptionListResult, err error) {
	req, err := client.ListPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "List", nil, "Failure preparing request")
	}

	resp, err := client.ListSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "List", resp, "Failure sending request")
	}

	result, err = client.ListResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "List", resp, "Failure responding to request")
	}

	return
}

// ListPreparer prepares the List request.
func (client ManagementClient) ListPreparer() (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions"),
		autorest.WithQueryParameters(queryParameters))
}

// ListSender sends the List request. The method will close the
// http.Response Body if it receives an error.
func (client ManagementClient) ListSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// ListResponder handles the response to the List request. The method always
// closes the http.Response Body.
func (client ManagementClient) ListResponder(resp *http.Response) (result SubscriptionListResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListNextResults retrieves the next set of results, if any.
func (client ManagementClient) ListNextResults(lastResults SubscriptionListResult) (result SubscriptionListResult, err error) {
	req, err := lastResults.SubscriptionListResultPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "List", nil, "Failure preparing next results request request")
	}
	if req == nil {
		return
	}

	resp, err := client.ListSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "List", resp, "Failure sending next results request request")
	}

	result, err = client.ListResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "List", resp, "Failure responding to next results request request")
	}

	return
}

// ListLocations gets a list of the subscription locations.
//
// subscriptionID is id of the subscription
func (client ManagementClient) ListLocations(subscriptionID string) (result LocationListResult, err error) {
	req, err := client.ListLocationsPreparer(subscriptionID)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "ListLocations", nil, "Failure preparing request")
	}

	resp, err := client.ListLocationsSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "ListLocations", resp, "Failure sending request")
	}

	result, err = client.ListLocationsResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "ListLocations", resp, "Failure responding to request")
	}

	return
}

// ListLocationsPreparer prepares the ListLocations request.
func (client ManagementClient) ListLocationsPreparer(subscriptionID string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(subscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/locations"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// ListLocationsSender sends the ListLocations request. The method will close the
// http.Response Body if it receives an error.
func (client ManagementClient) ListLocationsSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// ListLocationsResponder handles the response to the ListLocations request. The method always
// closes the http.Response Body.
func (client ManagementClient) ListLocationsResponder(resp *http.Response) (result LocationListResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListLocationsNextResults retrieves the next set of results, if any.
func (client ManagementClient) ListLocationsNextResults(lastResults LocationListResult) (result LocationListResult, err error) {
	req, err := lastResults.LocationListResultPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "ListLocations", nil, "Failure preparing next results request request")
	}
	if req == nil {
		return
	}

	resp, err := client.ListLocationsSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "ListLocations", resp, "Failure sending next results request request")
	}

	result, err = client.ListLocationsResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "subscriptions.ManagementClient", "ListLocations", resp, "Failure responding to next results request request")
	}

	return
}
