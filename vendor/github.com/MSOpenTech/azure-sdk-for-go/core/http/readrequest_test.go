// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"testing"
)

type reqTest struct {
	Raw     string
	Req     *Request
	Body    string
	Trailer Header
	Error   string
}

var noError = ""
var noBody = ""
var noTrailer Header = nil

var reqTests = []reqTest{
	// Baseline test; All Request fields included for template use
	{
		"GET http://www.techcrunch.com/ HTTP/1.1\r\n" +
			"Host: www.techcrunch.com\r\n" +
			"User-Agent: Fake\r\n" +
			"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
			"Accept-Language: en-us,en;q=0.5\r\n" +
			"Accept-Encoding: gzip,deflate\r\n" +
			"Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7\r\n" +
			"Keep-Alive: 300\r\n" +
			"Content-Length: 7\r\n" +
			"Proxy-Connection: keep-alive\r\n\r\n" +
			"abcdef\n???",

		&Request{
			Method: "GET",
			URL: &url.URL{
				Scheme: "http",
				Host:   "www.techcrunch.com",
				Path:   "/",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: Header{
				"Accept":           {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
				"Accept-Language":  {"en-us,en;q=0.5"},
				"Accept-Encoding":  {"gzip,deflate"},
				"Accept-Charset":   {"ISO-8859-1,utf-8;q=0.7,*;q=0.7"},
				"Keep-Alive":       {"300"},
				"Proxy-Connection": {"keep-alive"},
				"Content-Length":   {"7"},
				"User-Agent":       {"Fake"},
			},
			Close:         false,
			ContentLength: 7,
			Host:          "www.techcrunch.com",
			RequestURI:    "http://www.techcrunch.com/",
		},

		"abcdef\n",

		noTrailer,
		noError,
	},

	// GET request with no body (the normal case)
	{
		"GET / HTTP/1.1\r\n" +
			"Host: foo.com\r\n\r\n",

		&Request{
			Method: "GET",
			URL: &url.URL{
				Path: "/",
			},
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        Header{},
			Close:         false,
			ContentLength: 0,
			Host:          "foo.com",
			RequestURI:    "/",
		},

		noBody,
		noTrailer,
		noError,
	},

	// Tests that we don't parse a path that looks like a
	// scheme-relative URI as a scheme-relative URI.
	{
		"GET //user@host/is/actually/a/path/ HTTP/1.1\r\n" +
			"Host: test\r\n\r\n",

		&Request{
			Method: "GET",
			URL: &url.URL{
				Path: "//user@host/is/actually/a/path/",
			},
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        Header{},
			Close:         false,
			ContentLength: 0,
			Host:          "test",
			RequestURI:    "//user@host/is/actually/a/path/",
		},

		noBody,
		noTrailer,
		noError,
	},

	// Tests a bogus abs_path on the Request-Line (RFC 2616 section 5.1.2)
	{
		"GET ../../../../etc/passwd HTTP/1.1\r\n" +
			"Host: test\r\n\r\n",
		nil,
		noBody,
		noTrailer,
		"parse ../../../../etc/passwd: invalid URI for request",
	},

	// Tests missing URL:
	{
		"GET  HTTP/1.1\r\n" +
			"Host: test\r\n\r\n",
		nil,
		noBody,
		noTrailer,
		"parse : empty url",
	},

	// Tests chunked body with trailer:
	{
		"POST / HTTP/1.1\r\n" +
			"Host: foo.com\r\n" +
			"Transfer-Encoding: chunked\r\n\r\n" +
			"3\r\nfoo\r\n" +
			"3\r\nbar\r\n" +
			"0\r\n" +
			"Trailer-Key: Trailer-Value\r\n" +
			"\r\n",
		&Request{
			Method: "POST",
			URL: &url.URL{
				Path: "/",
			},
			TransferEncoding: []string{"chunked"},
			Proto:            "HTTP/1.1",
			ProtoMajor:       1,
			ProtoMinor:       1,
			Header:           Header{},
			ContentLength:    -1,
			Host:             "foo.com",
			RequestURI:       "/",
		},

		"foobar",
		Header{
			"Trailer-Key": {"Trailer-Value"},
		},
		noError,
	},

	// CONNECT request with domain name:
	{
		"CONNECT www.google.com:443 HTTP/1.1\r\n\r\n",

		&Request{
			Method: "CONNECT",
			URL: &url.URL{
				Host: "www.google.com:443",
			},
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        Header{},
			Close:         false,
			ContentLength: 0,
			Host:          "www.google.com:443",
			RequestURI:    "www.google.com:443",
		},

		noBody,
		noTrailer,
		noError,
	},

	// CONNECT request with IP address:
	{
		"CONNECT 127.0.0.1:6060 HTTP/1.1\r\n\r\n",

		&Request{
			Method: "CONNECT",
			URL: &url.URL{
				Host: "127.0.0.1:6060",
			},
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        Header{},
			Close:         false,
			ContentLength: 0,
			Host:          "127.0.0.1:6060",
			RequestURI:    "127.0.0.1:6060",
		},

		noBody,
		noTrailer,
		noError,
	},

	// CONNECT request for RPC:
	{
		"CONNECT /_goRPC_ HTTP/1.1\r\n\r\n",

		&Request{
			Method: "CONNECT",
			URL: &url.URL{
				Path: "/_goRPC_",
			},
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        Header{},
			Close:         false,
			ContentLength: 0,
			Host:          "",
			RequestURI:    "/_goRPC_",
		},

		noBody,
		noTrailer,
		noError,
	},

	// SSDP Notify request. golang.org/issue/3692
	{
		"NOTIFY * HTTP/1.1\r\nServer: foo\r\n\r\n",
		&Request{
			Method: "NOTIFY",
			URL: &url.URL{
				Path: "*",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: Header{
				"Server": []string{"foo"},
			},
			Close:         false,
			ContentLength: 0,
			RequestURI:    "*",
		},

		noBody,
		noTrailer,
		noError,
	},

	// OPTIONS request. Similar to golang.org/issue/3692
	{
		"OPTIONS * HTTP/1.1\r\nServer: foo\r\n\r\n",
		&Request{
			Method: "OPTIONS",
			URL: &url.URL{
				Path: "*",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: Header{
				"Server": []string{"foo"},
			},
			Close:         false,
			ContentLength: 0,
			RequestURI:    "*",
		},

		noBody,
		noTrailer,
		noError,
	},
}

func TestReadRequest(t *testing.T) {
	for i := range reqTests {
		tt := &reqTests[i]
		var braw bytes.Buffer
		braw.WriteString(tt.Raw)
		req, err := ReadRequest(bufio.NewReader(&braw))
		if err != nil {
			if err.Error() != tt.Error {
				t.Errorf("#%d: error %q, want error %q", i, err.Error(), tt.Error)
			}
			continue
		}
		rbody := req.Body
		req.Body = nil
		diff(t, fmt.Sprintf("#%d Request", i), req, tt.Req)
		var bout bytes.Buffer
		if rbody != nil {
			_, err := io.Copy(&bout, rbody)
			if err != nil {
				t.Fatalf("#%d. copying body: %v", i, err)
			}
			rbody.Close()
		}
		body := bout.String()
		if body != tt.Body {
			t.Errorf("#%d: Body = %q want %q", i, body, tt.Body)
		}
		if !reflect.DeepEqual(tt.Trailer, req.Trailer) {
			t.Errorf("#%d. Trailers differ.\n got: %v\nwant: %v", i, req.Trailer, tt.Trailer)
		}
	}
}
