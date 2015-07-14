package clcgo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var deletionResponse = "Deletion Response Body"

type testDeletionStatusProviding struct {
	CallbackForStatus func([]byte) (Status, error)
}

func (e testDeletionStatusProviding) URL(a string) (string, error) {
	return "/entity/url", nil
}

func (e testDeletionStatusProviding) StatusFromDeleteResponse(r []byte) (Status, error) {
	if e.CallbackForStatus != nil {
		return e.CallbackForStatus(r)
	}

	return Status{}, nil
}

func TestSuccessfulDeleteEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	e := testEntity{
		CallbackForURL: func(a string) (string, error) {
			assert.Equal(t, "AA", a)
			return "/entity/url", nil
		},
	}

	r.registerHandler("DELETE", "/entity/url", func(token string, req request) (string, error) {
		assert.Equal(t, "token", token)

		return deletionResponse, nil
	})

	status, err := c.DeleteEntity(&e)
	assert.NoError(t, err)
	assert.True(t, status.HasSucceeded())
}

func TestSuccessfulDeletionStatusProvidingEntityDeleteEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	st := Status{Status: "Custom"}
	e := testDeletionStatusProviding{
		CallbackForStatus: func(r []byte) (Status, error) {
			assert.Equal(t, []byte(deletionResponse), r)
			return st, nil
		},
	}

	r.registerHandler("DELETE", "/entity/url", func(token string, req request) (string, error) {
		return deletionResponse, nil
	})

	status, err := c.DeleteEntity(&e)
	assert.NoError(t, err)
	assert.Equal(t, st, status)
}

func TestErroredURLDeleteEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	e := testEntity{
		CallbackForURL: func(a string) (string, error) {
			return "", errors.New("test URL Error")
		},
	}

	status, err := c.DeleteEntity(&e)
	assert.EqualError(t, err, "test URL Error")
	assert.Equal(t, Status{}, status)
}

func TestErroredDeleteJSONDeleteEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	e := testEntity{}

	r.registerHandler("DELETE", "/entity/url", func(token string, req request) (string, error) {
		return "", errors.New("test DeleteJSON Error")
	})

	status, err := c.DeleteEntity(&e)
	assert.EqualError(t, err, "test DeleteJSON Error")
	assert.Equal(t, Status{}, status)
}

func TestDeletionStatusProvidingEntityDeleteEntity(t *testing.T) {
	r := newTestRequestor()
	cr := APICredentials{BearerToken: "token", AccountAlias: "AA"}
	c := Client{Requestor: r, APICredentials: cr}
	e := testDeletionStatusProviding{
		CallbackForStatus: func(r []byte) (Status, error) {
			return Status{}, errors.New("test StatusFromDeleteResponse Error")
		},
	}

	r.registerHandler("DELETE", "/entity/url", func(token string, req request) (string, error) {
		return deletionResponse, nil
	})

	status, err := c.DeleteEntity(&e)
	assert.EqualError(t, err, "test StatusFromDeleteResponse Error")
	assert.Equal(t, Status{}, status)
}
