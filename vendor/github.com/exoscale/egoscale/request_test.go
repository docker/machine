package egoscale

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPrepareValues(t *testing.T) {
	type tag struct {
		Name      string `json:"name"`
		IsVisible bool   `json:"isvisible,omitempty"`
	}

	profile := struct {
		IgnoreMe    string
		Zone        string            `json:"myzone,omitempty"`
		Name        string            `json:"name"`
		NoName      string            `json:"omitempty"`
		ID          int               `json:"id"`
		UserID      uint              `json:"user_id"`
		IsGreat     bool              `json:"is_great"`
		Num         float64           `json:"num"`
		Bytes       []byte            `json:"bytes"`
		IDs         []string          `json:"ids,omitempty"`
		TagPointers []*tag            `json:"tagpointers,omitempty"`
		Tags        []tag             `json:"tags,omitempty"`
		Map         map[string]string `json:"map"`
		IP          net.IP            `json:"ip,omitempty"`
	}{
		IgnoreMe: "bar",
		Name:     "world",
		NoName:   "foo",
		ID:       1,
		UserID:   uint(2),
		Num:      3.14,
		Bytes:    []byte("exo"),
		IDs:      []string{"1", "2", "three"},
		TagPointers: []*tag{
			{Name: "foo"},
			{Name: "bar", IsVisible: false},
		},
		Tags: []tag{
			{Name: "foo"},
			{Name: "bar", IsVisible: false},
		},
		Map: map[string]string{
			"foo": "bar",
		},
		IP: net.IPv4(192, 168, 0, 11),
	}

	params := url.Values{}
	err := prepareValues("", &params, profile)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := params["myzone"]; ok {
		t.Errorf("myzone params shouldn't be set, got %v", params.Get("myzone"))
	}

	if params.Get("NoName") != "foo" {
		t.Errorf("NoName params wasn't properly set, got %v", params.Get("NoName"))
	}

	if params.Get("name") != "world" {
		t.Errorf("name params wasn't properly set, got %v", params.Get("name"))
	}

	if params.Get("bytes") != "ZXhv" {
		t.Errorf("bytes params wasn't properly encoded in base 64, got %v", params.Get("bytes"))
	}

	if params.Get("ids") != "1,2,three" {
		t.Errorf("array of strings, wasn't property encoded, got %v", params.Get("ids"))
	}

	if _, ok := params["ignoreme"]; ok {
		t.Errorf("IgnoreMe key was set")
	}

	v := params.Get("tags[0].name")
	if v != "foo" {
		t.Errorf("expected tags to be serialized as foo, got %#v", v)
	}

	v = params.Get("tagpointers[0].name")
	if v != "foo" {
		t.Errorf("expected tag pointers to be serialized as foo, got %#v", v)
	}

	v = params.Get("map[0].foo")
	if v != "bar" {
		t.Errorf("expected map to be serialized as .foo => \"bar\", got .foo => %#v", v)
	}

	v = params.Get("is_great")
	if v != "false" {
		t.Errorf("expected bool to be serialized as \"false\", got %#v", v)
	}

	v = params.Get("ip")
	if v != "192.168.0.11" {
		t.Errorf("expected ip to be serialized as \"192.168.0.11\", got %#v", v)
	}
}

func TestPrepareValuesStringRequired(t *testing.T) {
	profile := struct {
		RequiredField string `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesBoolRequired(t *testing.T) {
	profile := struct {
		RequiredField bool `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err != nil {
		t.Fatal(nil)
	}
	if params.Get("requiredfield") != "false" {
		t.Errorf("bool params wasn't set to false (default value)")
	}
}

func TestPrepareValuesIntRequired(t *testing.T) {
	profile := struct {
		RequiredField int64 `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesUintRequired(t *testing.T) {
	profile := struct {
		RequiredField uint64 `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesBytesRequired(t *testing.T) {
	profile := struct {
		RequiredField []byte `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesSliceString(t *testing.T) {
	profile := struct {
		RequiredField []string `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesIP(t *testing.T) {
	profile := struct {
		RequiredField net.IP `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesIPZero(t *testing.T) {
	profile := struct {
		RequiredField net.IP `json:"requiredfield"`
	}{
		RequiredField: net.IPv4zero,
	}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestPrepareValuesMap(t *testing.T) {
	profile := struct {
		RequiredField map[string]string `json:"requiredfield"`
	}{}

	params := url.Values{}
	err := prepareValues("", &params, &profile)
	if err == nil {
		t.Errorf("It should have failed")
	}
}

func TestRequest(t *testing.T) {
	params := url.Values{}
	params.Set("command", "listApis")
	params.Set("token", "TOKEN")
	params.Set("name", "dummy")
	params.Set("response", "json")
	ts := newPostServer(params, `
{
	"listapisresponse": {
		"api": [{
			"name": "dummy",
			"description": "this is a test",
			"isasync": false,
			"since": "4.4",
			"type": "list",
			"name": "listDummies",
			"params": [],
			"related": "",
			"response": []
		}],
		"count": 1
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &ListAPIs{
		Name: "dummy",
	}
	resp, err := cs.Request(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	apis := resp.(*ListAPIsResponse)
	if apis.Count != 1 {
		t.Errorf("Expected exactly one API, got %d", apis.Count)
	}
}

func TestBooleanAsyncRequest(t *testing.T) {
	params := url.Values{}
	params.Set("command", "expungevirtualmachine")
	params.Set("token", "TOKEN")
	params.Set("id", "123")
	params.Set("response", "json")
	ts := newPostServer(params, `
{
	"expungevirtualmarchine": {
		"jobid": "1",
		"jobresult": {
			"success": true,
			"displaytext": "good job!"
		},
		"jobstatus": 1
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	req := &ExpungeVirtualMachine{
		ID: "123",
	}
	err := cs.BooleanAsyncRequest(req, AsyncInfo{})

	if err != nil {
		t.Errorf(err.Error())
	}
}

func newServer(code int, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Write([]byte(response))
	})
	return httptest.NewServer(mux)
}

func newPostServer(params url.Values, response string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		errors := make(map[string][]string)
		for k, expected := range params {
			if values, ok := (r.PostForm)[k]; ok {
				for i, value := range values {
					e := expected[i]
					if e != value {
						if _, ok := errors[k]; !ok {
							errors[k] = make([]string, len(values))
						}
						errors[k][i] = fmt.Sprintf("%s expected %v, got %v", k, e, value)
					}
				}
			}
		}

		if len(errors) == 0 {
			w.WriteHeader(200)
			w.Write([]byte(response))
		} else {
			w.WriteHeader(400)
			body, _ := json.Marshal(errors)
			w.Write(body)
			log.Println(body)
		}
	})
	return httptest.NewServer(mux)
}
