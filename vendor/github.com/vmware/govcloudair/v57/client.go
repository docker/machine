package v57

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// LoginPath for vCloud Air
	LoginPath = "/iam/login"
	// InstancesPath for vCloud Air
	InstancesPath = "/sc/instances"
	// HeaderAccept the HTTP accept header key
	HeaderAccept = "Accept"
	// JSONMimeV57 the json mime for version 5.7 of the API
	JSONMimeV57 = "application/json;version=5.7"
	// AnyXMLMime511 the wildcard xml mime for version 5.11 of the API
	AnyXMLMime511 = "application/*+xml;version=5.11"
	// Version511 the 5.11 version
	Version511 = "5.11"
	// Version is the default version number
	Version = Version511
)

// NewClient returns a new empty client to authenticate against the vCloud Air
// service, the vCloud Air endpoint can be overridden by setting the
// VCLOUDAIR_ENDPOINT environment variable.
func NewClient() (*Client, error) {
	var u *url.URL
	var err error

	if os.Getenv("VCLOUDAIR_ENDPOINT") != "" {
		u, err = url.ParseRequestURI(os.Getenv("VCLOUDAIR_ENDPOINT"))
		if err != nil {
			return nil, fmt.Errorf("cannot parse endpoint coming from VCLOUDAIR_ENDPOINT")
		}
	} else {
		// Implicitly trust this URL parse.
		u, _ = url.ParseRequestURI("https://vca.vmware.com/api")
	}
	return &Client{
		VAEndpoint:    *u,
		VCDAuthHeader: "X-Vcloud-Authorization",
		http:          http.Client{Transport: &http.Transport{TLSHandshakeTimeout: 120 * time.Second}},
	}, nil
}

// Client provides a client to vCloud Air, values can be populated automatically using the Authenticate method.
type Client struct {
	VAToken       string      // vCloud Air authorization token
	VAEndpoint    url.URL     // vCloud Air API endpoint
	Region        string      // Region where the compute resource lives.
	VCDToken      string      // Access Token (authorization header)
	VCDAuthHeader string      // Authorization header
	vcdHREF       *url.URL    // HREF of the backend VDC you're using
	http          http.Client // HttpClient is the client to use. Default will be used if not provided.
}

// Authenticate is a helper function that performs a complete login in vCloud
// Air and in the backend vCloud Director instance.
func (c *Client) Authenticate(username, password string) error {
	if username == "" {
		username = os.Getenv("VCLOUDAIR_USERNAME")
	}
	if password == "" {
		password = os.Getenv("VCLOUDAIR_PASSWORD")
	}

	r, _ := http.NewRequest("POST", c.VAEndpoint.String()+LoginPath, nil)
	r.Header.Set("Accept", JSONMimeV57)
	r.SetBasicAuth(username, password)

	resp, err := c.http.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Could not complete request with vca, because (status %d) %s\n", resp.StatusCode, resp.Status)
	}

	var result oAuthClient
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return err
	}
	result.AuthToken = resp.Header.Get("vchs-authorization")
	c.VAToken = result.AuthToken
	result.Config = c

	instances, err := result.instances()
	if err != nil {
		return err
	}

	var attrs *accountInstanceAttrs
	for _, inst := range instances {
		attrs = inst.Attrs()
		if attrs != nil {
			c.Region = inst.Region
			break
		}
	}

	if attrs == nil {
		return fmt.Errorf("unable to determine session url")
	}
	attrs.client = c

	return attrs.Authenticate(username, password)
}

// Disconnect performs a disconnection from the vCloud Air API endpoint.
func (c *Client) Disconnect() error {
	return nil
}

// NewRequest creates a new HTTP request and applies necessary auth headers if
// set.
func (c *Client) NewRequest(params map[string]string, method string, u *url.URL, body io.Reader) *http.Request {

	p := url.Values{}

	// Build up our request parameters
	for k, v := range params {
		p.Add(k, v)
	}

	// Add the params to our URL
	u.RawQuery = p.Encode()

	// Build the request, no point in checking for errors here as we're just
	// passing a string version of an url.URL struct and http.NewRequest returns
	// error only if can't process an url.ParseRequestURI().
	req, _ := http.NewRequest(method, u.String(), body)

	if c.VCDToken != "" {
		// Add the authorization header
		req.Header.Add(c.VCDAuthHeader, c.VCDToken)
		// Add the Accept header for VCD
		req.Header.Add("Accept", "application/*+xml;version="+Version)
	}
	return req

}

// NewAuthenticatedSession create a new vCloud Air authenticated client
func NewAuthenticatedSession(user, password string) (*Client, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	if err := client.Authenticate(user, password); err != nil {
		return nil, err
	}

	return client, nil
}

type oAuthClient struct {
	AuthToken       string   `json:"-"`
	Config          *Client  `json:"-"`
	ServiceGroupIDs []string `json:"serviceGroupIds"`
	Info            struct {
		Instances []accountInstance `json:"instances"`
	}
}

func (a *oAuthClient) instances() ([]accountInstance, error) {
	if err := a.JSONRequest("GET", InstancesPath, &a.Info); err != nil {
		return nil, err
	}
	return a.Info.Instances, nil
}

func (a *oAuthClient) JSONRequest(method, path string, result interface{}) error {
	r, _ := http.NewRequest(method, a.Config.VAEndpoint.String()+path, nil)
	r.Header.Set(HeaderAccept, JSONMimeV57)

	if a.AuthToken != "" {
		r.Header.Set("Authorization", "Bearer "+a.AuthToken)
	}

	resp, err := a.Config.http.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Could not complete request with vca, because (status %d) %s\n", resp.StatusCode, resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(result); err != nil {
		return err
	}

	return nil
}

type accountInstance struct {
	APIURL             string   `json:"apiUrl"`
	DashboardURL       string   `json:"dashboardUrl"`
	Description        string   `json:"description"`
	ID                 string   `json:"id"`
	InstanceAttributes string   `json:"instanceAttributes"`
	InstanceVersion    string   `json:"instanceVersion"`
	Link               []string `json:"link"`
	Name               string   `json:"name"`
	PlanID             string   `json:"planId"`
	Region             string   `json:"region"`
	ServiceGroupID     string   `json:"serviceGroupId"`
}

func (a *accountInstance) Attrs() *accountInstanceAttrs {
	if !strings.HasPrefix(a.InstanceAttributes, "{") {
		return nil
	}

	var res accountInstanceAttrs
	if err := json.Unmarshal([]byte(a.InstanceAttributes), &res); err != nil {
		return nil
	}
	return &res
}

type accountInstanceAttrs struct {
	OrgName       string `json:"orgName"`
	SessionURI    string `json:"sessionUri"`
	APIVersionURI string `json:"apiVersionUri"`
	client        *Client
}

func (a *accountInstanceAttrs) Authenticate(user, password string) error {
	r, _ := http.NewRequest("POST", a.SessionURI, nil)
	r.Header.Set(HeaderAccept, AnyXMLMime511)
	r.SetBasicAuth(user+"@"+a.OrgName, password)

	resp, err := a.client.http.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Could not complete authenticating with vCloud, because (status %d) %s\n", resp.StatusCode, resp.Status)
	}
	a.client.VCDToken = resp.Header.Get("x-vcloud-authorization")

	return nil
}
