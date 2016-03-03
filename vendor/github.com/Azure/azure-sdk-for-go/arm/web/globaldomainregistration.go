package web

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

// GlobalDomainRegistrationClient is the use these APIs to manage Azure
// Websites resources through the Azure Resource Manager. All task operations
// conform to the HTTP/1.1 protocol specification and each operation returns
// an x-ms-request-id header that can be used to obtain information about the
// request. You must make sure that requests made to these resources are
// secure. For more information, see <a
// href="https://msdn.microsoft.com/en-us/library/azure/dn790557.aspx">Authenticating
// Azure Resource Manager requests.</a>
type GlobalDomainRegistrationClient struct {
	ManagementClient
}

// NewGlobalDomainRegistrationClient creates an instance of the
// GlobalDomainRegistrationClient client.
func NewGlobalDomainRegistrationClient(subscriptionID string) GlobalDomainRegistrationClient {
	return NewGlobalDomainRegistrationClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewGlobalDomainRegistrationClientWithBaseURI creates an instance of the
// GlobalDomainRegistrationClient client.
func NewGlobalDomainRegistrationClientWithBaseURI(baseURI string, subscriptionID string) GlobalDomainRegistrationClient {
	return GlobalDomainRegistrationClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// CheckDomainAvailability sends the check domain availability request.
//
// identifier is name of the domain
func (client GlobalDomainRegistrationClient) CheckDomainAvailability(identifier NameIdentifier) (result DomainAvailablilityCheckResult, ae error) {
	req, err := client.CheckDomainAvailabilityPreparer(identifier)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "CheckDomainAvailability", nil, "Failure preparing request")
	}

	resp, err := client.CheckDomainAvailabilitySender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "CheckDomainAvailability", resp, "Failure sending request")
	}

	result, err = client.CheckDomainAvailabilityResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "CheckDomainAvailability", resp, "Failure responding to request")
	}

	return
}

// CheckDomainAvailabilityPreparer prepares the CheckDomainAvailability request.
func (client GlobalDomainRegistrationClient) CheckDomainAvailabilityPreparer(identifier NameIdentifier) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/providers/Microsoft.DomainRegistration/checkDomainAvailability"),
		autorest.WithJSON(identifier),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// CheckDomainAvailabilitySender sends the CheckDomainAvailability request. The method will close the
// http.Response Body if it receives an error.
func (client GlobalDomainRegistrationClient) CheckDomainAvailabilitySender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// CheckDomainAvailabilityResponder handles the response to the CheckDomainAvailability request. The method always
// closes the http.Response Body.
func (client GlobalDomainRegistrationClient) CheckDomainAvailabilityResponder(resp *http.Response) (result DomainAvailablilityCheckResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// GetAllDomains sends the get all domains request.
func (client GlobalDomainRegistrationClient) GetAllDomains() (result DomainCollection, ae error) {
	req, err := client.GetAllDomainsPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "GetAllDomains", nil, "Failure preparing request")
	}

	resp, err := client.GetAllDomainsSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "GetAllDomains", resp, "Failure sending request")
	}

	result, err = client.GetAllDomainsResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "GetAllDomains", resp, "Failure responding to request")
	}

	return
}

// GetAllDomainsPreparer prepares the GetAllDomains request.
func (client GlobalDomainRegistrationClient) GetAllDomainsPreparer() (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/providers/Microsoft.DomainRegistration/domains"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetAllDomainsSender sends the GetAllDomains request. The method will close the
// http.Response Body if it receives an error.
func (client GlobalDomainRegistrationClient) GetAllDomainsSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// GetAllDomainsResponder handles the response to the GetAllDomains request. The method always
// closes the http.Response Body.
func (client GlobalDomainRegistrationClient) GetAllDomainsResponder(resp *http.Response) (result DomainCollection, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// GetDomainControlCenterSsoRequest sends the get domain control center sso
// request request.
func (client GlobalDomainRegistrationClient) GetDomainControlCenterSsoRequest() (result DomainControlCenterSsoRequest, ae error) {
	req, err := client.GetDomainControlCenterSsoRequestPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "GetDomainControlCenterSsoRequest", nil, "Failure preparing request")
	}

	resp, err := client.GetDomainControlCenterSsoRequestSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "GetDomainControlCenterSsoRequest", resp, "Failure sending request")
	}

	result, err = client.GetDomainControlCenterSsoRequestResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "GetDomainControlCenterSsoRequest", resp, "Failure responding to request")
	}

	return
}

// GetDomainControlCenterSsoRequestPreparer prepares the GetDomainControlCenterSsoRequest request.
func (client GlobalDomainRegistrationClient) GetDomainControlCenterSsoRequestPreparer() (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/providers/Microsoft.DomainRegistration/generateSsoRequest"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetDomainControlCenterSsoRequestSender sends the GetDomainControlCenterSsoRequest request. The method will close the
// http.Response Body if it receives an error.
func (client GlobalDomainRegistrationClient) GetDomainControlCenterSsoRequestSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// GetDomainControlCenterSsoRequestResponder handles the response to the GetDomainControlCenterSsoRequest request. The method always
// closes the http.Response Body.
func (client GlobalDomainRegistrationClient) GetDomainControlCenterSsoRequestResponder(resp *http.Response) (result DomainControlCenterSsoRequest, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListDomainRecommendations sends the list domain recommendations request.
//
// parameters is domain recommendation search parameters
func (client GlobalDomainRegistrationClient) ListDomainRecommendations(parameters DomainRecommendationSearchParameters) (result NameIdentifierCollection, ae error) {
	req, err := client.ListDomainRecommendationsPreparer(parameters)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "ListDomainRecommendations", nil, "Failure preparing request")
	}

	resp, err := client.ListDomainRecommendationsSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "ListDomainRecommendations", resp, "Failure sending request")
	}

	result, err = client.ListDomainRecommendationsResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "ListDomainRecommendations", resp, "Failure responding to request")
	}

	return
}

// ListDomainRecommendationsPreparer prepares the ListDomainRecommendations request.
func (client GlobalDomainRegistrationClient) ListDomainRecommendationsPreparer(parameters DomainRecommendationSearchParameters) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/providers/Microsoft.DomainRegistration/listDomainRecommendations"),
		autorest.WithJSON(parameters),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// ListDomainRecommendationsSender sends the ListDomainRecommendations request. The method will close the
// http.Response Body if it receives an error.
func (client GlobalDomainRegistrationClient) ListDomainRecommendationsSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// ListDomainRecommendationsResponder handles the response to the ListDomainRecommendations request. The method always
// closes the http.Response Body.
func (client GlobalDomainRegistrationClient) ListDomainRecommendationsResponder(resp *http.Response) (result NameIdentifierCollection, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ValidateDomainPurchaseInformation sends the validate domain purchase
// information request.
//
// domainRegistrationInput is domain registration information
func (client GlobalDomainRegistrationClient) ValidateDomainPurchaseInformation(domainRegistrationInput DomainRegistrationInput) (result ObjectSet, ae error) {
	req, err := client.ValidateDomainPurchaseInformationPreparer(domainRegistrationInput)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "ValidateDomainPurchaseInformation", nil, "Failure preparing request")
	}

	resp, err := client.ValidateDomainPurchaseInformationSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "ValidateDomainPurchaseInformation", resp, "Failure sending request")
	}

	result, err = client.ValidateDomainPurchaseInformationResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/GlobalDomainRegistrationClient", "ValidateDomainPurchaseInformation", resp, "Failure responding to request")
	}

	return
}

// ValidateDomainPurchaseInformationPreparer prepares the ValidateDomainPurchaseInformation request.
func (client GlobalDomainRegistrationClient) ValidateDomainPurchaseInformationPreparer(domainRegistrationInput DomainRegistrationInput) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": url.QueryEscape(client.SubscriptionID),
	}

	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	return autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPath("/subscriptions/{subscriptionId}/providers/Microsoft.DomainRegistration/validateDomainRegistrationInformation"),
		autorest.WithJSON(domainRegistrationInput),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// ValidateDomainPurchaseInformationSender sends the ValidateDomainPurchaseInformation request. The method will close the
// http.Response Body if it receives an error.
func (client GlobalDomainRegistrationClient) ValidateDomainPurchaseInformationSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// ValidateDomainPurchaseInformationResponder handles the response to the ValidateDomainPurchaseInformation request. The method always
// closes the http.Response Body.
func (client GlobalDomainRegistrationClient) ValidateDomainPurchaseInformationResponder(resp *http.Response) (result ObjectSet, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Value),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
