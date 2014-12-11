package gopherstack

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type CloudstackClient struct {
	// The http client for communicating
	client *http.Client

	// The base URL of the API
	BaseURL string

	// Credentials
	APIKey    string
	SecretKey string
}

// Creates a new client for communicating with Cloudstack
func (cloudstack CloudstackClient) New(apiurl string, apikey string, secretkey string, insecureskipverify bool) *CloudstackClient {
	c := &CloudstackClient{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureskipverify},
				Proxy:           http.ProxyFromEnvironment,
			},
		},
		BaseURL:   apiurl,
		APIKey:    apikey,
		SecretKey: secretkey,
	}
	return c
}

func NewRequest(c CloudstackClient, request string, params url.Values) (interface{}, error) {
	client := c.client

	params.Set("apikey", c.APIKey)
	params.Set("command", request)
	params.Set("response", "json")

	// Generate signature for API call
	// * Serialize parameters and sort them by key, done by Encode()
	// * Use byte sequence for '+' character as Cloudstack requires this
	// * For the signature only, un-encode [ and ].
	// * Convert the entire argument string to lowercase
	// * Calculate HMAC SHA1 of argument string with Cloudstack secret key
	// * URL encode the string and convert to base64
	s := params.Encode()
	s2 := strings.Replace(s, "+", "%20", -1)
	s3 := strings.ToLower(strings.Replace(strings.Replace(s2, "%5B", "[", -1), "%5D", "]", -1))
	mac := hmac.New(sha1.New, []byte(c.SecretKey))
	mac.Write([]byte(s3))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	signature = url.QueryEscape(signature)

	// Create the final URL before we issue the request
	// For some reason Cloudstack refuses to accept '+' as a space character so we byte escape it instead.
	url := c.BaseURL + "?" + s2 + "&signature=" + signature

	log.Printf("Calling %s ", url)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	log.Printf("Response from Cloudstack: %d - %s", resp.StatusCode, body)
	if resp.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("Received HTTP client/server error from Cloudstack: %d - %s", resp.StatusCode, body))
		return nil, err
	}

	switch request {
	default:
		log.Printf("Unknown request %s", request)
	case "createSSHKeyPair":
		var decodedResponse CreateSshKeyPairResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "deleteSSHKeyPair":
		var decodedResponse DeleteSshKeyPairResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "deployVirtualMachine":
		var decodedResponse DeployVirtualMachineResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "destroyVirtualMachine":
		var decodedResponse DestroyVirtualMachineResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "stopVirtualMachine":
		var decodedResponse StopVirtualMachineResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "listVirtualMachines":
		var decodedResponse ListVirtualMachinesResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "listProjects":
		var decodedResponse ListProjectsResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "listVolumes":
		var decodedResponse ListVolumesResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "createTemplate":
		var decodedResponse CreateTemplateResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "listTemplates":
		var decodedResponse ListTemplatesResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "listDiskOfferings":
		var decodedResponse ListDiskOfferingsResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "detachIso":
		var decodedResponse DetachIsoResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "queryAsyncJobResult":
		var decodedResponse QueryAsyncJobResultResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "createTags":
		var decodedResponse CreateTagsResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "listTags":
		var decodedResponse ListTagsResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	case "deleteTags":
		var decodedResponse DeleteTagsResponse
		json.Unmarshal(body, &decodedResponse)
		return decodedResponse, nil

	}

	// only reached with unknown request
	return "", nil
}
