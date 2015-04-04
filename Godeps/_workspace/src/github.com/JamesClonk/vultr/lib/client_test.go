package lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestServerAndClient(code int, body string) (*httptest.Server, *Client) {
	server := getTestServer(code, body)
	return server, getTestClient(server.URL)
}

func getTestServer(code int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
}

func getTestClient(endpoint string) *Client {
	options := Options{
		Endpoint: endpoint,
	}
	return NewClient("test-key", &options)
}

func Test_Client_Options(t *testing.T) {
	options := Options{
		HTTPClient: http.DefaultClient,
		UserAgent:  "test-agent",
		Endpoint:   "http://test",
	}

	client := NewClient("test-key", &options)
	if assert.NotNil(t, client) {
		assert.Equal(t, "test-key", client.APIKey)
		assert.Equal(t, "http://test", client.Endpoint.String())
		assert.Equal(t, "test-agent", client.UserAgent)
	}
}

func Test_Client_NewClient(t *testing.T) {
	client := NewClient("test-key-new", nil)
	if assert.NotNil(t, client) {
		assert.Equal(t, "test-key-new", client.APIKey)
		assert.Equal(t, "https://api.vultr.com/", client.Endpoint.String())
		assert.Equal(t, "vultr-go/"+Version, client.UserAgent)
	}
}
