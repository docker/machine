package clcgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CenturyLinkLabs/clcgo/fakeapi"
	"github.com/stretchr/testify/assert"
)

type testParameters struct {
	TestKey string
}

func TestSuccessfulUnauthenticatedPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Empty(t, r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("accepts"))

		s, _ := ioutil.ReadAll(r.Body)
		var p testParameters
		err := json.Unmarshal(s, &p)
		assert.NoError(t, err)
		assert.Equal(t, testParameters{"Testing"}, p)

		fmt.Fprintf(w, "Response Text")
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: testParameters{"Testing"}}
	response, err := r.PostJSON("", req)
	assert.NoError(t, err)

	responseString := string(response)
	assert.Equal(t, "Response Text", responseString)
}

func TestSuccessfulAuthenticatedPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		fmt.Fprintf(w, "Response Text")
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: testParameters{}}
	_, err := r.PostJSON("token", req)
	assert.NoError(t, err)
}

func TestSuccessfulPassingArrayToPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, `["first","second"]`, s)

		fmt.Fprintf(w, "Response Text")
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: []string{"first", "second"}}
	_, err := r.PostJSON("token", req)
	assert.NoError(t, err)
}

func TestUnauthorizedPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", 401)
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: testParameters{}}
	_, err := r.PostJSON("token", req)
	reqErr, ok := err.(RequestError)
	if assert.True(t, ok) {
		assert.EqualError(t, reqErr, "your bearer token was rejected")
		assert.Equal(t, 401, reqErr.StatusCode)
	}
}

func TestUnhandledStatusOnPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "I'm a teapot", 418)
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: testParameters{"Testing"}}
	response, err := r.PostJSON("", req)
	assert.Contains(t, string(response), "I'm a teapot")

	reqErr, ok := err.(RequestError)
	if assert.True(t, ok) {
		assert.EqualError(t, reqErr, "got an unexpected status code '418'")
		assert.Equal(t, 418, reqErr.StatusCode)
	}
}

func Test400WithMessagesOnPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fakeapi.ServerCreationInvalidResponse, 400)
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: testParameters{}}
	_, err := r.PostJSON("token", req)

	reqErr, ok := err.(RequestError)
	if assert.True(t, ok) {
		assert.EqualError(t, reqErr, "the request is invalid.")
		assert.Equal(t, 400, reqErr.StatusCode)
		if assert.Len(t, reqErr.Errors, 2) {
			assert.Equal(t, "The name field is required.", reqErr.Errors["body.name"][0])
			assert.Equal(
				t, "The sourceServerId field is required.", reqErr.Errors["body.sourceServerId"][0],
			)
		}
	}
}

func TestSuccessfulGetJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response Text")

		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("accepts"))
	}))
	defer ts.Close()

	r := &clcRequestor{}
	response, err := r.GetJSON("token", request{URL: ts.URL})
	assert.NoError(t, err)
	assert.Equal(t, "Response Text", string(response))
}

func TestUnauthorizedGetJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", 401)
	}))
	defer ts.Close()

	r := &clcRequestor{}
	_, err := r.GetJSON("token", request{URL: ts.URL})
	reqErr, ok := err.(RequestError)
	if assert.True(t, ok) {
		assert.EqualError(t, reqErr, "your bearer token was rejected")
		assert.Equal(t, 401, reqErr.StatusCode)
	}
}

func TestErrored400GetJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	}))
	defer ts.Close()

	r := &clcRequestor{}
	response, err := r.GetJSON("token", request{URL: ts.URL})
	assert.Contains(t, string(response), "Bad Request")

	reqErr, ok := err.(RequestError)
	if assert.True(t, ok) {
		assert.EqualError(t, reqErr, "got an unexpected status code '400'")
		assert.Equal(t, 400, reqErr.StatusCode)
	}
}

func TestSuccessfulDeleteJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("accepts"))

		w.WriteHeader(202)
		fmt.Fprintf(w, "Response Text")
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: testParameters{}}
	res, err := r.DeleteJSON("token", req)

	assert.NoError(t, err)
	assert.Equal(t, "Response Text", string(res))
}

func TestUnauthorizedDeleteJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", 401)
	}))
	defer ts.Close()

	r := &clcRequestor{}
	req := request{URL: ts.URL, Parameters: testParameters{}}
	_, err := r.DeleteJSON("token", req)
	reqErr, ok := err.(RequestError)
	if assert.True(t, ok) {
		assert.EqualError(t, reqErr, "your bearer token was rejected")
		assert.Equal(t, 401, reqErr.StatusCode)
	}
}

func TestErrored400DeleteJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	}))
	defer ts.Close()

	r := &clcRequestor{}
	response, err := r.DeleteJSON("token", request{URL: ts.URL})
	assert.Contains(t, string(response), "Bad Request")

	reqErr, ok := err.(RequestError)
	if assert.True(t, ok) {
		assert.EqualError(t, reqErr, "got an unexpected status code '400'")
		assert.Equal(t, 400, reqErr.StatusCode)
	}
}

func TestTypeFromLinks(t *testing.T) {
	l := Link{Rel: "t"}
	ls := []Link{l}

	found, err := typeFromLinks("t", ls)
	assert.NoError(t, err)
	assert.Equal(t, l, found)

	found, err = typeFromLinks("bad", ls)
	assert.EqualError(t, err, "no link of type 'bad'")
	assert.Equal(t, Link{}, found)
}
