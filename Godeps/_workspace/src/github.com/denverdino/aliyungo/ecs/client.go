package ecs

import (
	"bytes"
	"encoding/json"
	"github.com/denverdino/aliyungo/util"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// ECSDefaultEndpoint is the default API endpoint of ECS services
const ECSDefaultEndpoint = "https://ecs.aliyuncs.com"

// Interval for checking status in WaitForXXX method
const DefaultWaitForInterval = 5

// Default timeout value for WaitForXXX method
const DefaultTimeout = 60

// A Client represents a client of ECS services
type Client struct {
	AccessKeyId     string //Access Key Id
	AccessKeySecret string //Access Key Secret
	debug           bool
	httpClient      *http.Client
}

// NewClient creates a new instance of ECS client
func NewClient(accessKeyId, accessKeySecret string) *Client {
	client := &Client{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret + "&",
		debug:           false,
		httpClient:      &http.Client{},
	}
	return client
}

// SetAccessKeyId sets new AccessKeyId
func (client *Client) SetAccessKeyId(id string) {
	client.AccessKeyId = id
}

// SetAccessKeySecret sets new AccessKeySecret
func (client *Client) SetAccessKeySecret(secret string) {
	client.AccessKeySecret = secret + "&"
}

func getECSErrorFromString(str string) error {
	return &Error{
		ErrorResponse: ErrorResponse{
			Code:    "ECSClientFailure",
			Message: str,
		},
		StatusCode: -1,
	}
}

func getECSError(err error) error {
	return getECSErrorFromString(err.Error())
}

// SetDebug sets debug mode to log the request/response message
func (client *Client) SetDebug(debug bool) {
	client.debug = debug
}

// Invoke sends the raw HTTP request for ECS services
func (client *Client) Invoke(action string, args interface{}, response interface{}) error {

	request := Request{}
	request.init(action, client.AccessKeyId)

	query := util.ConvertToQueryValues(request)
	util.SetQueryValues(args, &query)

	// Sign request
	signature := util.CreateSignatureForRequest(ECSRequestMethod, &query, client.AccessKeySecret)

	// Generate the request URL
	requestURL := ECSDefaultEndpoint + "?" + query.Encode() + "&Signature=" + url.QueryEscape(signature)

	httpReq, err := http.NewRequest(ECSRequestMethod, requestURL, nil)
	if err != nil {
		return getECSError(err)
	}

	t0 := time.Now()
	httpResp, err := client.httpClient.Do(httpReq)
	t1 := time.Now()
	if err != nil {
		return getECSError(err)
	}
	statusCode := httpResp.StatusCode

	if client.debug {
		log.Printf("Invoke %s %s %d (%v)", ECSRequestMethod, requestURL, statusCode, t1.Sub(t0))
	}

	defer httpResp.Body.Close()
	body, err := ioutil.ReadAll(httpResp.Body)

	if err != nil {
		return getECSError(err)
	}

	if client.debug {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, body, "", "    ")
		log.Println(string(prettyJSON.Bytes()))
	}

	if statusCode >= 400 && statusCode <= 599 {
		errorResponse := ErrorResponse{}
		err = json.Unmarshal(body, &errorResponse)
		ecsError := &Error{
			ErrorResponse: errorResponse,
			StatusCode:    statusCode,
		}
		return ecsError
	}

	err = json.Unmarshal(body, response)
	//log.Printf("%++v", response)
	if err != nil {
		return getECSError(err)
	}

	return nil
}

// GenerateClientToken generates the Client Token with random string
func (client *Client) GenerateClientToken() string {
	return util.CreateRandomString()
}
