package clcgo_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/CenturyLinkLabs/clcgo"
	"github.com/CenturyLinkLabs/clcgo/fakeapi"
)

var (
	exampleDefaultHTTPClient *http.Client
	fakeAPIServer            *httptest.Server
)

type exampleDomainRewriter struct {
	RewriteURL *url.URL
}

func (r exampleDomainRewriter) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req.URL.Scheme = "http"
	req.URL.Host = r.RewriteURL.Host

	t := http.Transport{}
	return t.RoundTrip(req)
}

func setupExample() {
	fakeAPIServer = fakeapi.CreateFakeServer()

	// Replace the clcgo.DefaultHTTPClient with one that will rewrite the
	// requests to go to the test server instead of production.
	exampleDefaultHTTPClient = clcgo.DefaultHTTPClient
	u, _ := url.Parse(fakeAPIServer.URL)
	clcgo.DefaultHTTPClient = &http.Client{Transport: exampleDomainRewriter{RewriteURL: u}}
}

func teardownExample() {
	fakeAPIServer.Close()
	clcgo.DefaultHTTPClient = exampleDefaultHTTPClient
}

func ExampleClient_GetAPICredentials_successful() {
	// Some test-related setup code which you can safely ignore.
	setupExample()
	defer teardownExample()

	c := clcgo.NewClient()
	c.GetAPICredentials("user", "pass")

	fmt.Printf("Account Alias: %s", c.APICredentials.AccountAlias)
	// Output:
	// Account Alias: ACME
}

func ExampleClient_GetAPICredentials_failed() {
	// Some test-related setup code which you can safely ignore.
	setupExample()
	defer teardownExample()

	c := clcgo.NewClient()
	err := c.GetAPICredentials("bad", "bad")

	fmt.Printf("Error: %s", err)
	// Output:
	// Error: there was a problem with your credentials
}

func ExampleClient_GetEntity_successful() {
	// Some test-related setup code which you can safely ignore.
	setupExample()
	defer teardownExample()

	c := clcgo.NewClient()
	c.GetAPICredentials("user", "pass")

	s := clcgo.Server{ID: "server1"}
	c.GetEntity(&s)

	fmt.Printf("Server Name: %s", s.Name)
	// Output:
	// Server Name: Test Name
}

func ExampleClient_GetEntity_expiredToken() {
	// Some test-related setup code which you can safely ignore.
	setupExample()
	defer teardownExample()

	c := clcgo.NewClient()
	// You are caching this Bearer Token value and it has either expired or for
	// some other reason become invalid.
	c.APICredentials = clcgo.APICredentials{BearerToken: "expired", AccountAlias: "ACME"}

	s := clcgo.Server{ID: "server1"}
	err := c.GetEntity(&s)

	rerr, _ := err.(clcgo.RequestError)
	fmt.Printf("Error: %s, Status Code: %d", rerr, rerr.StatusCode)
	// Output:
	// Error: your bearer token was rejected, Status Code: 401
}

func ExampleClient_SaveEntity_successful() {
	// Some test-related setup code which you can safely ignore.
	setupExample()
	defer teardownExample()

	c := clcgo.NewClient()
	c.GetAPICredentials("user", "pass")

	// Build a Server resource. In reality there are many more required fields.
	s := clcgo.Server{Name: "My Server"}

	// Request the Server be provisioned, returning a Status. In your code you
	// must NOT ignore the possibility of an error here.
	st, _ := c.SaveEntity(&s)

	// Refresh the Status until it has completed. In your code you should put a
	// delay between requests, as this can take a while.
	if !st.HasSucceeded() {
		c.GetEntity(&st)
	}

	// The Status says that the server is provisioned. You can now request its
	// details. Again, your code should not ignore errors as this is doing.
	c.GetEntity(&s)

	fmt.Printf("Server ID: %s", s.ID)
	// Output:
	// Server ID: test-id
}

func ExampleCredentials_persisted() {
	client := clcgo.NewClient()
	creds := clcgo.APICredentials{BearerToken: "TOKEN", AccountAlias: "ACME"}
	client.APICredentials = creds // Client now ready for authenticated requests.
}
