package linodego

import (
	"encoding/json"
	"net/url"
)

// Test Service
type TestService struct {
	client *Client
}

// Test Service Response
type TestResponse struct {
	Response
	Data map[string]string
}

// Echo request with the given key and value
func (t *TestService) Echo(key string, val string, v *TestResponse) error {
	u := &url.Values{}
	u.Add(key, val)
	if err := t.client.do("test.echo", u, &v.Response); err != nil {
		return err
	}
	v.Data = map[string]string{}
	if err := json.Unmarshal(v.RawData, &v.Data); err != nil {
		return err
	}
	return nil
}
