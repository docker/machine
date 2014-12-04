package azureSdkForGo

import (
	"bytes"
	"crypto/rand"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/MSOpenTech/azure-sdk-for-go/core/http"
	"github.com/MSOpenTech/azure-sdk-for-go/core/tls"
	"io"
	"os/exec"
	"strings"
	"time"
)

const (
	ParamNotSpecifiedError = "Parameter %s is not specified."

	azureManagementDnsName = "https://management.core.windows.net"
	msVersionHeader        = "x-ms-version"
	msVersionHeaderValue   = "2014-05-01"
	contentHeader          = "Content-Type"
	contentHeaderValue     = "application/xml"
	requestIdHeader        = "X-Ms-Request-Id"
)

//Region public methods starts

func SendAzureGetRequest(url string) ([]byte, error) {
	if len(url) == 0 {
		return nil, fmt.Errorf(ParamNotSpecifiedError, "url")
	}

	response, err := SendAzureRequest(url, "GET", nil)
	if err != nil {
		return nil, err
	}

	responseContent := getResponseBody(response)
	return responseContent, nil
}

func SendAzurePostRequest(url string, data []byte) (string, error) {
	if len(url) == 0 {
		return "", fmt.Errorf(ParamNotSpecifiedError, "url")
	}

	response, err := SendAzureRequest(url, "POST", data)
	if err != nil {
		return "", err
	}

	requestId := response.Header[requestIdHeader]
	return requestId[0], nil
}

func SendAzureDeleteRequest(url string) (string, error) {
	if len(url) == 0 {
		return "", fmt.Errorf(ParamNotSpecifiedError, "url")
	}

	response, err := SendAzureRequest(url, "DELETE", nil)
	if err != nil {
		return "", err
	}

	requestId := response.Header[requestIdHeader]
	return requestId[0], nil
}

func SendAzureRequest(url string, requestType string, data []byte) (*http.Response, error) {
	if len(url) == 0 {
		return nil, fmt.Errorf(ParamNotSpecifiedError, "url")
	}
	if len(requestType) == 0 {
		return nil, fmt.Errorf(ParamNotSpecifiedError, "requestType")
	}

	client := createHttpClient()

	response, err := sendRequest(client, url, requestType, data, 7)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func ExecuteCommand(command string, input []byte) ([]byte, error) {
	if len(command) == 0 {
		return nil, fmt.Errorf(ParamNotSpecifiedError, "command")
	}

	parts := strings.Fields(command)
	head := parts[0]
	parts = parts[1:len(parts)]

	cmd := exec.Command(head, parts...)
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func GetOperationStatus(operationId string) (*Operation, error) {
	if len(operationId) == 0 {
		return nil, fmt.Errorf(ParamNotSpecifiedError, "operationId")
	}

	operation := new(Operation)
	url := "operations/" + operationId
	response, azureErr := SendAzureGetRequest(url)
	if azureErr != nil {
		return nil, azureErr
	}

	err := xml.Unmarshal(response, operation)
	if err != nil {
		return nil, err
	}

	return operation, nil
}

func WaitAsyncOperation(operationId string) error {
	if len(operationId) == 0 {
		return fmt.Errorf(ParamNotSpecifiedError, "operationId")
	}

	status := "InProgress"
	operation := new(Operation)
	err := errors.New("")
	for status == "InProgress" {
		time.Sleep(2000 * time.Millisecond)
		operation, err = GetOperationStatus(operationId)
		if err != nil {
			return err
		}

		status = operation.Status
	}

	if status == "Failed" {
		return errors.New(operation.Error.Message)
	}

	return nil
}

func CheckStringParams(url string) ([]byte, error) {
	if len(url) == 0 {
		return nil, fmt.Errorf(ParamNotSpecifiedError, "url")
	}

	response, err := SendAzureRequest(url, "GET", nil)
	if err != nil {
		return nil, err
	}

	responseContent := getResponseBody(response)
	return responseContent, nil
}

// NewUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40

	//return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
	return fmt.Sprintf("%x", uuid[10:]), nil
}

//Region public methods ends

//Region private methods starts

func sendRequest(client *http.Client, url string, requestType string, data []byte, numberOfRetries int) (*http.Response, error) {
	request, reqErr := createAzureRequest(url, requestType, data)
	if reqErr != nil {
		return nil, reqErr
	}

	response, err := client.Do(request)
	if err != nil {
		if numberOfRetries == 0 {
			return nil, err
		}

		return sendRequest(client, url, requestType, data, numberOfRetries-1)
	}

	if response.StatusCode > 299 {
		responseContent := getResponseBody(response)
		azureErr := getAzureError(responseContent)
		if azureErr != nil {
			if numberOfRetries == 0 {
				return nil, azureErr
			}

			return sendRequest(client, url, requestType, data, numberOfRetries-1)
		}
	}

	return response, nil
}

func getAzureError(responseBody []byte) error {
	error := new(AzureError)
	err := xml.Unmarshal(responseBody, error)
	if err != nil {
		return err
	}

	return error
}

func createAzureRequest(url string, requestType string, data []byte) (*http.Request, error) {
	var request *http.Request
	var err error

	url = fmt.Sprintf("%s/%s/%s", azureManagementDnsName, GetPublishSettings().SubscriptionID, url)
	if data != nil {
		body := bytes.NewBuffer(data)
		request, err = http.NewRequest(requestType, url, body)
	} else {
		request, err = http.NewRequest(requestType, url, nil)
	}

	if err != nil {
		return nil, err
	}

	request.Header.Add(msVersionHeader, msVersionHeaderValue)
	request.Header.Add(contentHeader, contentHeaderValue)

	return request, nil
}

func createHttpClient() *http.Client {
	cert, _ := tls.X509KeyPair(GetPublishSettings().SubscriptionCert, GetPublishSettings().SubscriptionKey)

	ssl := &tls.Config{}
	ssl.Certificates = []tls.Certificate{cert}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: ssl,
		},
	}

	return client
}

func getResponseBody(response *http.Response) []byte {

	responseBody := make([]byte, response.ContentLength)
	io.ReadFull(response.Body, responseBody)
	return responseBody
}

//Region private methods ends

type AzureError struct {
	XMLName xml.Name `xml:"Error"`
	Code    string
	Message string
}

func (e *AzureError) Error() string {
	return fmt.Sprintf("Code: %s, Message: %s", e.Code, e.Message)
}

type Operation struct {
	XMLName        xml.Name `xml:"Operation"`
	ID             string
	Status         string
	HttpStatusCode string
	Error          AzureError
}
