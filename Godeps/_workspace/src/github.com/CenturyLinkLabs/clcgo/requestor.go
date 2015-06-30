package clcgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// DefaultHTTPClient can be overridden in the same fashion as the DefaultClient
// for http, if you would like to implement some behavioral change around the
// requests that we could not anticipate.
var DefaultHTTPClient = &http.Client{}

type requestor interface {
	PostJSON(string, request) ([]byte, error)
	GetJSON(string, request) ([]byte, error)
	DeleteJSON(string, request) ([]byte, error)
}

type clcRequestor struct{}

type modelStates map[string][]string

// A RequestError can be returned from GetEntity and SaveEntity calls and
// contain specific information about the unexpected error from your request.
// It can be especially helpful on SaveEntity validation failures, for instance
// if you omitted a required field.
type RequestError struct {
	Message    string
	StatusCode int
	Errors     modelStates
}

type invalidReqestResponse struct {
	Message    string      `json:"message"`
	ModelState modelStates `json:"modelState"`
}

func (r RequestError) Error() string {
	return r.Message
}

func (r clcRequestor) PostJSON(t string, req request) ([]byte, error) {
	body, code, err := requestFor("POST", t, req)
	if err != nil {
		return []byte{}, err
	}

	switch code {
	case 200, 201, 202:
		return body, nil
	case 400:
		var e invalidReqestResponse
		err := json.Unmarshal(body, &e)
		if err != nil {
			return body, err
		}

		return body, RequestError{Message: e.Message, StatusCode: 400, Errors: e.ModelState}
	case 401:
		return body, RequestError{Message: "your bearer token was rejected", StatusCode: 401}
	default:
		return body, RequestError{Message: fmt.Sprintf("got an unexpected status code '%d'", code), StatusCode: code}
	}
}

func (r clcRequestor) GetJSON(t string, req request) ([]byte, error) {
	body, code, err := requestFor("GET", t, req)
	if err != nil {
		return []byte{}, err
	}

	switch code {
	case 200:
		return body, nil
	case 401:
		return body, RequestError{Message: "your bearer token was rejected", StatusCode: 401}
	default:
		return body, RequestError{Message: fmt.Sprintf("got an unexpected status code '%d'", code), StatusCode: code}
	}
}

func (r clcRequestor) DeleteJSON(t string, req request) ([]byte, error) {
	body, code, err := requestFor("DELETE", t, req)
	if err != nil {
		return []byte{}, err
	}

	switch code {
	case 202:
		return body, nil
	case 401:
		return body, RequestError{Message: "your bearer token was rejected", StatusCode: 401}
	default:
		return body, RequestError{Message: fmt.Sprintf("got an unexpected status code '%d'", code), StatusCode: code}
	}
}

func requestFor(m string, t string, req request) ([]byte, int, error) {
	var params io.Reader
	if req.Parameters != nil {
		j, err := json.Marshal(req.Parameters)
		if err != nil {
			return []byte{}, 0, err
		}

		params = bytes.NewReader(j)
	}

	hr, err := http.NewRequest(m, req.URL, params)
	if err != nil {
		return []byte{}, 0, err
	}

	if t != "" {
		hr.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t))
	}
	hr.Header.Add("Content-Type", "application/json")
	hr.Header.Add("Accepts", "application/json")

	resp, err := DefaultHTTPClient.Do(hr)
	if err != nil {
		return []byte{}, 0, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, err
	}

	return body, resp.StatusCode, nil
}

func typeFromLinks(t string, ls []Link) (Link, error) {
	for _, l := range ls {
		if l.Rel == t {
			return l, nil
		}
	}

	return Link{}, fmt.Errorf("no link of type '%s'", t)
}
