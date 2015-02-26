package linodego

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	// API Client Version
	ClientVersion = "0.1"
	// Default Base URL
	defaultBaseURL = "https://api.linode.com"
)

var (
	// Unexpected Response Error
	ErrUnexpectedResponse = errors.New("Unexpected response")
	// General Response Error
	ErrResponse = errors.New("Error response")
	// Authentication Failure
	ErrAuthentication = errors.New("Authenticaion Failed. Be sure to use correct API Key")
)

// Client of Linode v1 API
type Client struct {
	// Linode API Key
	ApiKey string
	// HTTP client to communicate with Linode API
	HTTPClient *http.Client
	// Base URL
	BaseURL *url.URL

	// Whether to use POST for API request, default is false
	UsePost bool

	// Services
	Test    *TestService
	Api     *ApiService
	Avail   *AvailService
	Account *AccountService
	Image   *ImageService
	Linode  *LinodeService
	Job     *LinodeJobService
	Config  *LinodeConfigService
	Ip      *LinodeIPService
	Disk    *LinodeDiskService
}

// Creates a new Linode client object.
func NewClient(AccessKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel) // set debug level
	}

	c := &Client{ApiKey: AccessKey, HTTPClient: httpClient, BaseURL: baseURL}
	c.Test = &TestService{client: c}
	c.Api = &ApiService{client: c}
	c.Avail = &AvailService{client: c}
	c.Account = &AccountService{client: c}
	c.Image = &ImageService{client: c}
	c.Linode = &LinodeService{client: c}
	c.Job = &LinodeJobService{client: c}
	c.Config = &LinodeConfigService{client: c}
	c.Ip = &LinodeIPService{client: c}
	c.Job = &LinodeJobService{client: c}
	c.Disk = &LinodeDiskService{client: c}
	return c
}

// execute request
func (c *Client) do(action string, params *url.Values, v *Response) error {
	// https://api.linode.com/?api_key={key}&api_action=test.echo&foo=bar
	if params == nil {
		params = &url.Values{}
	}
	params.Add("api_key", c.ApiKey)
	params.Add("api_action", action)
	return c.request(params, v)
}

// send request via POST to Linode API. The response is stored in the value pointed to by v
// Returns an error if an API error has occurred.
func (c *Client) request(params *url.Values, v *Response) error {
	var body string
	if params != nil {
		body = params.Encode()
	}

	method := "GET"
	requestURL := fmt.Sprintf("%v/?%s", c.BaseURL, body)
	var requestBody io.Reader
	if c.UsePost {
		method = "POST"
		requestURL = c.BaseURL.String()
		requestBody = strings.NewReader(body)
	}
	request, err := http.NewRequest(method, requestURL, requestBody)
	if err != nil {
		return err
	}

	log.Debugf("HTTP REQUEST: %s %s %s", method, requestURL, body)

	if c.UsePost {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("User-Agent", "LinodeGo/"+ClientVersion+" Go/"+runtime.Version())

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		log.Errorf("Failed to get API response: ", err)
		return err
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Failed to parse API response: ", err)
		return err
	}

	log.Debugf("HTTP RESPONSE: %s", string(responseBody))

	// Status code 500 is a server error and means nothing can be done at this point.
	if response.StatusCode != 200 {
		return ErrUnexpectedResponse
	}

	if err = json.Unmarshal(responseBody, v); err != nil {
		return err
	}

	if len(v.Errors) > 0 {
		if v.Errors[0].ErrorCode == 4 {
			return ErrAuthentication
		}
		// Need to return all errors to client. Perhaps wrap the error structure as a property for a custom error class
		return errors.New(fmt.Sprintf("API Error: %s, Code %d", v.Errors[0].ErrorMessage, v.Errors[0].ErrorCode))
	}

	return nil
}
