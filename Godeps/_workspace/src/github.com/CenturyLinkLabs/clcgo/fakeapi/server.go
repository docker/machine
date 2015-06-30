package fakeapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// CreateFakeServer instantiates a new httptest.Server hooked the the correct
// routes and returning the expected fixture data.
func CreateFakeServer() *httptest.Server {
	m := http.NewServeMux()
	s := httptest.NewServer(m)

	// Authentication
	m.Handle("/v2/authentication/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := struct {
			Username string
			Password string
		}{}
		s, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(s, &p)

		if p.Username == "user" && p.Password == "pass" {
			fmt.Fprintf(w, AuthenticationSuccessfulResponse)
		} else {
			http.Error(w, "{}", 400)
		}
	}))

	// Server fetch
	m.Handle("/v2/servers/ACME/server1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer 1234ABCDEF" {
			http.Error(w, "", 401)
		} else {
			fmt.Fprintf(w, ServerResponse)
		}
	}))

	// Server save
	m.Handle("/v2/servers/ACME", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer 1234ABCDEF" {
			http.Error(w, "", 401)
		} else if r.Method != "POST" {
			http.Error(w, "", 405)
		} else {
			fmt.Fprintf(w, ServerCreationSuccessfulResponse)
		}
	}))

	// Status fetch
	m.Handle("/v2/operations/alias/status/test-status-id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer 1234ABCDEF" {
			http.Error(w, "", 401)
		} else {
			fmt.Fprintf(w, SuccessfulStatusResponse)
		}
	}))

	// Server fetch by UUID
	m.Handle("/v2/servers/alias/test-uuid", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer 1234ABCDEF" {
			http.Error(w, "", 401)
		} else {
			fmt.Fprintf(w, ServerResponse)
		}
	}))
	return s
}
