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

// CertificatesClient is the use these APIs to manage Azure Websites resources
// through the Azure Resource Manager. All task operations conform to the
// HTTP/1.1 protocol specification and each operation returns an
// x-ms-request-id header that can be used to obtain information about the
// request. You must make sure that requests made to these resources are
// secure. For more information, see <a
// href="https://msdn.microsoft.com/en-us/library/azure/dn790557.aspx">Authenticating
// Azure Resource Manager requests.</a>
type CertificatesClient struct {
	ManagementClient
}

// NewCertificatesClient creates an instance of the CertificatesClient client.
func NewCertificatesClient(subscriptionID string) CertificatesClient {
	return NewCertificatesClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewCertificatesClientWithBaseURI creates an instance of the
// CertificatesClient client.
func NewCertificatesClientWithBaseURI(baseURI string, subscriptionID string) CertificatesClient {
	return CertificatesClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// CreateOrUpdateCertificate sends the create or update certificate request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate. certificateEnvelope is details of certificate if it exists
// already.
func (client CertificatesClient) CreateOrUpdateCertificate(resourceGroupName string, name string, certificateEnvelope Certificate) (result Certificate, ae error) {
	req, err := client.CreateOrUpdateCertificatePreparer(resourceGroupName, name, certificateEnvelope)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "CreateOrUpdateCertificate", nil, "Failure preparing request")
	}

	resp, err := client.CreateOrUpdateCertificateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "CreateOrUpdateCertificate", resp, "Failure sending request")
	}

	result, err = client.CreateOrUpdateCertificateResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "CreateOrUpdateCertificate", resp, "Failure responding to request")
	}

	return
}

// CreateOrUpdateCertificatePreparer prepares the CreateOrUpdateCertificate request.
func (client CertificatesClient) CreateOrUpdateCertificatePreparer(resourceGroupName string, name string, certificateEnvelope Certificate) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/certificates/{name}"),
		autorest.WithJSON(certificateEnvelope),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// CreateOrUpdateCertificateSender sends the CreateOrUpdateCertificate request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) CreateOrUpdateCertificateSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// CreateOrUpdateCertificateResponder handles the response to the CreateOrUpdateCertificate request. The method always
// closes the http.Response Body.
func (client CertificatesClient) CreateOrUpdateCertificateResponder(resp *http.Response) (result Certificate, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// CreateOrUpdateCsr sends the create or update csr request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate. csrEnvelope is details of certificate signing request if it
// exists already.
func (client CertificatesClient) CreateOrUpdateCsr(resourceGroupName string, name string, csrEnvelope Csr) (result Csr, ae error) {
	req, err := client.CreateOrUpdateCsrPreparer(resourceGroupName, name, csrEnvelope)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "CreateOrUpdateCsr", nil, "Failure preparing request")
	}

	resp, err := client.CreateOrUpdateCsrSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "CreateOrUpdateCsr", resp, "Failure sending request")
	}

	result, err = client.CreateOrUpdateCsrResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "CreateOrUpdateCsr", resp, "Failure responding to request")
	}

	return
}

// CreateOrUpdateCsrPreparer prepares the CreateOrUpdateCsr request.
func (client CertificatesClient) CreateOrUpdateCsrPreparer(resourceGroupName string, name string, csrEnvelope Csr) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/csrs/{name}"),
		autorest.WithJSON(csrEnvelope),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// CreateOrUpdateCsrSender sends the CreateOrUpdateCsr request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) CreateOrUpdateCsrSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// CreateOrUpdateCsrResponder handles the response to the CreateOrUpdateCsr request. The method always
// closes the http.Response Body.
func (client CertificatesClient) CreateOrUpdateCsrResponder(resp *http.Response) (result Csr, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// DeleteCertificate sends the delete certificate request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate to be deleted.
func (client CertificatesClient) DeleteCertificate(resourceGroupName string, name string) (result ObjectSet, ae error) {
	req, err := client.DeleteCertificatePreparer(resourceGroupName, name)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "DeleteCertificate", nil, "Failure preparing request")
	}

	resp, err := client.DeleteCertificateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "DeleteCertificate", resp, "Failure sending request")
	}

	result, err = client.DeleteCertificateResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "DeleteCertificate", resp, "Failure responding to request")
	}

	return
}

// DeleteCertificatePreparer prepares the DeleteCertificate request.
func (client CertificatesClient) DeleteCertificatePreparer(resourceGroupName string, name string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/certificates/{name}"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// DeleteCertificateSender sends the DeleteCertificate request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) DeleteCertificateSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// DeleteCertificateResponder handles the response to the DeleteCertificate request. The method always
// closes the http.Response Body.
func (client CertificatesClient) DeleteCertificateResponder(resp *http.Response) (result ObjectSet, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Value),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// DeleteCsr sends the delete csr request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate signing request.
func (client CertificatesClient) DeleteCsr(resourceGroupName string, name string) (result ObjectSet, ae error) {
	req, err := client.DeleteCsrPreparer(resourceGroupName, name)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "DeleteCsr", nil, "Failure preparing request")
	}

	resp, err := client.DeleteCsrSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "DeleteCsr", resp, "Failure sending request")
	}

	result, err = client.DeleteCsrResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "DeleteCsr", resp, "Failure responding to request")
	}

	return
}

// DeleteCsrPreparer prepares the DeleteCsr request.
func (client CertificatesClient) DeleteCsrPreparer(resourceGroupName string, name string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/csrs/{name}"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// DeleteCsrSender sends the DeleteCsr request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) DeleteCsrSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// DeleteCsrResponder handles the response to the DeleteCsr request. The method always
// closes the http.Response Body.
func (client CertificatesClient) DeleteCsrResponder(resp *http.Response) (result ObjectSet, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Value),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// GetCertificate sends the get certificate request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate.
func (client CertificatesClient) GetCertificate(resourceGroupName string, name string) (result Certificate, ae error) {
	req, err := client.GetCertificatePreparer(resourceGroupName, name)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCertificate", nil, "Failure preparing request")
	}

	resp, err := client.GetCertificateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCertificate", resp, "Failure sending request")
	}

	result, err = client.GetCertificateResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCertificate", resp, "Failure responding to request")
	}

	return
}

// GetCertificatePreparer prepares the GetCertificate request.
func (client CertificatesClient) GetCertificatePreparer(resourceGroupName string, name string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/certificates/{name}"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetCertificateSender sends the GetCertificate request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) GetCertificateSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// GetCertificateResponder handles the response to the GetCertificate request. The method always
// closes the http.Response Body.
func (client CertificatesClient) GetCertificateResponder(resp *http.Response) (result Certificate, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// GetCertificates sends the get certificates request.
//
// resourceGroupName is name of the resource group
func (client CertificatesClient) GetCertificates(resourceGroupName string) (result CertificateCollection, ae error) {
	req, err := client.GetCertificatesPreparer(resourceGroupName)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCertificates", nil, "Failure preparing request")
	}

	resp, err := client.GetCertificatesSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCertificates", resp, "Failure sending request")
	}

	result, err = client.GetCertificatesResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCertificates", resp, "Failure responding to request")
	}

	return
}

// GetCertificatesPreparer prepares the GetCertificates request.
func (client CertificatesClient) GetCertificatesPreparer(resourceGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/certificates"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetCertificatesSender sends the GetCertificates request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) GetCertificatesSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// GetCertificatesResponder handles the response to the GetCertificates request. The method always
// closes the http.Response Body.
func (client CertificatesClient) GetCertificatesResponder(resp *http.Response) (result CertificateCollection, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// GetCsr sends the get csr request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate.
func (client CertificatesClient) GetCsr(resourceGroupName string, name string) (result Csr, ae error) {
	req, err := client.GetCsrPreparer(resourceGroupName, name)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCsr", nil, "Failure preparing request")
	}

	resp, err := client.GetCsrSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCsr", resp, "Failure sending request")
	}

	result, err = client.GetCsrResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCsr", resp, "Failure responding to request")
	}

	return
}

// GetCsrPreparer prepares the GetCsr request.
func (client CertificatesClient) GetCsrPreparer(resourceGroupName string, name string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/csrs/{name}"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetCsrSender sends the GetCsr request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) GetCsrSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// GetCsrResponder handles the response to the GetCsr request. The method always
// closes the http.Response Body.
func (client CertificatesClient) GetCsrResponder(resp *http.Response) (result Csr, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// GetCsrs sends the get csrs request.
//
// resourceGroupName is name of the resource group
func (client CertificatesClient) GetCsrs(resourceGroupName string) (result CsrList, ae error) {
	req, err := client.GetCsrsPreparer(resourceGroupName)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCsrs", nil, "Failure preparing request")
	}

	resp, err := client.GetCsrsSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCsrs", resp, "Failure sending request")
	}

	result, err = client.GetCsrsResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "GetCsrs", resp, "Failure responding to request")
	}

	return
}

// GetCsrsPreparer prepares the GetCsrs request.
func (client CertificatesClient) GetCsrsPreparer(resourceGroupName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/csrs"),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// GetCsrsSender sends the GetCsrs request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) GetCsrsSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// GetCsrsResponder handles the response to the GetCsrs request. The method always
// closes the http.Response Body.
func (client CertificatesClient) GetCsrsResponder(resp *http.Response) (result CsrList, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Value),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// UpdateCertificate sends the update certificate request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate. certificateEnvelope is details of certificate if it exists
// already.
func (client CertificatesClient) UpdateCertificate(resourceGroupName string, name string, certificateEnvelope Certificate) (result Certificate, ae error) {
	req, err := client.UpdateCertificatePreparer(resourceGroupName, name, certificateEnvelope)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "UpdateCertificate", nil, "Failure preparing request")
	}

	resp, err := client.UpdateCertificateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "UpdateCertificate", resp, "Failure sending request")
	}

	result, err = client.UpdateCertificateResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "UpdateCertificate", resp, "Failure responding to request")
	}

	return
}

// UpdateCertificatePreparer prepares the UpdateCertificate request.
func (client CertificatesClient) UpdateCertificatePreparer(resourceGroupName string, name string, certificateEnvelope Certificate) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/certificates/{name}"),
		autorest.WithJSON(certificateEnvelope),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// UpdateCertificateSender sends the UpdateCertificate request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) UpdateCertificateSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// UpdateCertificateResponder handles the response to the UpdateCertificate request. The method always
// closes the http.Response Body.
func (client CertificatesClient) UpdateCertificateResponder(resp *http.Response) (result Certificate, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// UpdateCsr sends the update csr request.
//
// resourceGroupName is name of the resource group name is name of the
// certificate. csrEnvelope is details of certificate signing request if it
// exists already.
func (client CertificatesClient) UpdateCsr(resourceGroupName string, name string, csrEnvelope Csr) (result Csr, ae error) {
	req, err := client.UpdateCsrPreparer(resourceGroupName, name, csrEnvelope)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "UpdateCsr", nil, "Failure preparing request")
	}

	resp, err := client.UpdateCsrSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "web/CertificatesClient", "UpdateCsr", resp, "Failure sending request")
	}

	result, err = client.UpdateCsrResponder(resp)
	if err != nil {
		ae = autorest.NewErrorWithError(err, "web/CertificatesClient", "UpdateCsr", resp, "Failure responding to request")
	}

	return
}

// UpdateCsrPreparer prepares the UpdateCsr request.
func (client CertificatesClient) UpdateCsrPreparer(resourceGroupName string, name string, csrEnvelope Csr) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"name":              url.QueryEscape(name),
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
		autorest.WithPath("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Web/csrs/{name}"),
		autorest.WithJSON(csrEnvelope),
		autorest.WithPathParameters(pathParameters),
		autorest.WithQueryParameters(queryParameters))
}

// UpdateCsrSender sends the UpdateCsr request. The method will close the
// http.Response Body if it receives an error.
func (client CertificatesClient) UpdateCsrSender(req *http.Request) (*http.Response, error) {
	return client.Send(req)
}

// UpdateCsrResponder handles the response to the UpdateCsr request. The method always
// closes the http.Response Body.
func (client CertificatesClient) UpdateCsrResponder(resp *http.Response) (result Csr, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
