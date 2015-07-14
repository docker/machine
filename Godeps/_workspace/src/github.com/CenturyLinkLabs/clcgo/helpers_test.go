package clcgo

import "fmt"

type handlerCallback func(string, request) (string, error)

type testRequestor struct {
	Handlers map[string]map[string]handlerCallback
}

func newTestRequestor() testRequestor {
	return testRequestor{
		Handlers: make(map[string]map[string]handlerCallback),
	}
}

// BUG(dp): Handlers that are never called will not fail. There needs to be
// some kind of ability to verify that a handler was called if you wish.
func (r *testRequestor) registerHandler(m string, url string, callback handlerCallback) {
	if _, found := r.Handlers[m]; !found {
		r.Handlers[m] = make(map[string]handlerCallback)
	}

	r.Handlers[m][url] = callback
}

func (r testRequestor) GetJSON(t string, req request) ([]byte, error) {
	return r.responseForMethod("GET", t, req)
}

func (r testRequestor) PostJSON(t string, req request) ([]byte, error) {
	return r.responseForMethod("POST", t, req)
}

func (r testRequestor) DeleteJSON(t string, req request) ([]byte, error) {
	return r.responseForMethod("DELETE", t, req)
}

func (r testRequestor) responseForMethod(m string, t string, req request) ([]byte, error) {
	callback, found := r.Handlers[m][req.URL]
	if found {
		s, err := callback(t, req)
		return []byte(s), err
	}

	return nil, fmt.Errorf("there is no handler for %s '%s'", m, req.URL)
}
