package egoscale

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"
)

// Gettable represents an Interface that can be "Get" by the client
type Gettable interface {
	// Get populates the given resource or throws
	Get(context context.Context, client *Client) error
}

// Deletable represents an Interface that can be "Delete" by the client
type Deletable interface {
	// Delete removes the given resources or throws
	Delete(context context.Context, client *Client) error
}

// Client represents the CloudStack API client
type Client struct {
	client    *http.Client
	endpoint  string
	apiKey    string
	apiSecret string
	// Timeout represents the default timeout for the async requests
	Timeout time.Duration
	// RetryStrategy represents the waiting strategy for polling the async requests
	RetryStrategy RetryStrategyFunc
}

// Get populates the given resource or fails
func (client *Client) Get(g Gettable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return g.Get(ctx, client)
}

// GetWithContext populates the given resource or fails
func (client *Client) GetWithContext(ctx context.Context, g Gettable) error {
	return g.Get(ctx, client)
}

// Delete removes the given resource of fails
func (client *Client) Delete(g Deletable) error {
	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	return g.Delete(ctx, client)
}

// DeleteWithContext removes the given resource of fails
func (client *Client) DeleteWithContext(ctx context.Context, g Deletable) error {
	return g.Delete(ctx, client)
}

// RetryStrategyFunc represents a how much time to wait between two calls to CloudStack
type RetryStrategyFunc func(int64) time.Duration

// NewClientWithTimeout creates a CloudStack API client
//
// Timeout is set to booth the HTTP client and the client itself.
func NewClientWithTimeout(endpoint, apiKey, apiSecret string, timeout time.Duration) *Client {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	cs := &Client{
		client:        client,
		endpoint:      endpoint,
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		Timeout:       timeout,
		RetryStrategy: FibonacciRetryStrategy,
	}

	return cs
}

// NewClient creates a CloudStack API client with default timeout (60)
func NewClient(endpoint, apiKey, apiSecret string) *Client {
	timeout := time.Duration(60 * time.Second)
	return NewClientWithTimeout(endpoint, apiKey, apiSecret, timeout)
}

// FibonacciRetryStrategy waits for an increasing amount of time following the Fibonacci sequence
func FibonacciRetryStrategy(iteration int64) time.Duration {
	var a, b, i, tmp int64
	a = 0
	b = 1
	for i = 0; i < iteration; i++ {
		tmp = a + b
		a = b
		b = tmp
	}
	return time.Duration(a) * time.Second
}
