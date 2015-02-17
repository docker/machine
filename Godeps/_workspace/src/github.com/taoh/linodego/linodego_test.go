package linodego

import (
	"net/url"
	"testing"
)

func TestApiRequest(t *testing.T) {
	client := NewClient(APIKey, nil)
	u := &url.Values{}
	u.Add("foo", "bar")
	v := &Response{}
	if err := client.do("test.echo", u, v); err != nil {
		t.Fatal(err)
	}
}

func TestErrorApiRequest(t *testing.T) {
	client := NewClient(APIKey, nil)
	u := &url.Values{}
	v := &Response{}
	err := client.do("test.non_implemented_method", u, v)
	if err == nil {
		t.Fatal("Should not succeed in executing method not implemented")
	}
}
