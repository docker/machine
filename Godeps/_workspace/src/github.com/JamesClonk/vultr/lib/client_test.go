package lib

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		Endpoint:       endpoint,
		RateLimitation: 1 * time.Millisecond,
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

// Test that API queries are throttled to 1/sec
func Test_Client_Throttling(t *testing.T) {
	const ERROR = 250 * time.Millisecond
	const EXPECTED_DURATION = 4 * time.Second
	server, _ := getTestServerAndClient(http.StatusOK, `{
		"balance":-15.97,"pending_charges":"2.34",
		"last_payment_date":"2015-01-29 05:06:27","last_payment_amount":"-5.00"}`)
	defer server.Close()

	options := Options{
		Endpoint:       server.URL,
		RateLimitation: 1 * time.Second,
	}
	client := NewClient("test-key", &options)

	time.Sleep(1 * time.Second)

	// The first query should not be throttled
	info, _ := client.GetAccountInfo()

	// The next four queries should be throttled and take 4 seconds
	before := time.Now()
	info, _ = client.GetAccountInfo()
	info, _ = client.GetAccountInfo()
	info, _ = client.GetAccountInfo()
	info, _ = client.GetAccountInfo()
	after := time.Now()

	lower := EXPECTED_DURATION - ERROR
	upper := EXPECTED_DURATION + ERROR
	assert.NotNil(t, info)
	if diff := after.Sub(before); diff < lower || diff > upper {
		t.Errorf("Waited %s seconds, though really should have waited between %s and %s", diff.String(), lower.String(), upper.String())
	}
}
