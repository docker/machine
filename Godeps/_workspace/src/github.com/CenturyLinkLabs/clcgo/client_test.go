package clcgo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var entityResponse = `{"testSerializedKey":"value"}`

type testEntity struct {
	CallbackForURL    func(string) (string, error)
	TestSerializedKey string `json:"testSerializedKey"`
}

func (e testEntity) URL(a string) (string, error) {
	if e.CallbackForURL != nil {
		return e.CallbackForURL(a)
	}

	return "/entity/url", nil
}

func TestSuccessfulGetEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}

	r.registerHandler("GET", "/entity", func(token string, req request) (string, error) {
		assert.Equal(t, "token", token)
		return entityResponse, nil
	})

	e := testEntity{
		CallbackForURL: func(a string) (string, error) {
			assert.Equal(t, "AA", a)
			return "/entity", nil
		},
	}

	err := c.GetEntity(&e)
	assert.NoError(t, err)
	assert.Equal(t, "value", e.TestSerializedKey)
}

func TestErroredURLInGetEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	e := testEntity{
		CallbackForURL: func(a string) (string, error) {
			return "", errors.New("test URL Error")
		},
	}

	err := c.GetEntity(&e)
	assert.EqualError(t, err, "test URL Error")
}

func TestErroredInGetJSONInGetEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	e := testEntity{}
	r.registerHandler("GET", "/entity/url", func(token string, req request) (string, error) {
		return "", errors.New("error from GetJSON")
	})

	err := c.GetEntity(&e)
	assert.EqualError(t, err, "error from GetJSON")
}

func TestBadJSONInGetJSONInGetEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	e := testEntity{}
	r.registerHandler("GET", "/entity/url", func(token string, req request) (string, error) {
		return ``, nil
	})

	err := c.GetEntity(&e)
	assert.EqualError(t, err, "unexpected end of JSON input")
}
