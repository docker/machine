// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package rest

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Request struct {
	Method      string
	URL         string
	ContentType string
	Body        io.Reader
	Token       string
}

type Page struct {
	Items            []interface{} `json:"items"`
	NextPageLink     string        `json:"nextPageLink"`
	PreviousPageLink string        `json:"previousPageLink"`
}

type DocumentList struct {
	Items []interface{}
}

const appJson string = "application/json"

func AppendSlice(origSlice []interface{}, dataToAppend []interface{}) []interface{} {
	origLen := len(origSlice)
	newLen := origLen + len(dataToAppend)

	if newLen > cap(origSlice) {
		newSlice := make([]interface{}, (newLen+1)*2)
		copy(newSlice, origSlice)
		origSlice = newSlice
	}

	origSlice = origSlice[0:newLen]
	copy(origSlice[origLen:newLen], dataToAppend)

	return origSlice
}

func Get(client *http.Client, url string, token string) (res *http.Response, err error) {
	req := Request{"GET", url, "", nil, token}
	res, err = Do(client, &req)
	return
}

func GetList(client *http.Client, endpoint string, url string, token string) (result []byte, err error) {
	req := Request{"GET", url, "", nil, token}
	res, err := Do(client, &req)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		errMsg := fmt.Sprintf("photon: HTTP %d: %v", res.StatusCode, res.Body)
		return nil, errors.New(errMsg)
	}

	decoder := json.NewDecoder(res.Body)
	decoder.UseNumber()

	page := &Page{}
	err = decoder.Decode(page)
	if err != nil {
		return
	}

	documentList := &DocumentList{}
	documentList.Items = AppendSlice(documentList.Items, page.Items)

	for page.NextPageLink != "" {
		req = Request{"GET", endpoint + page.NextPageLink, "", nil, token}
		res, err = Do(client, &req)
		if err != nil {
			return
		}

		decoder = json.NewDecoder(res.Body)
		decoder.UseNumber()

		page.NextPageLink = ""
		page.PreviousPageLink = ""

		err = decoder.Decode(page)
		if err != nil {
			return
		}

		documentList.Items = AppendSlice(documentList.Items, page.Items)
	}

	result, err = json.Marshal(documentList)

	return
}

func Post(client *http.Client, url string, contentType string, body io.Reader, token string) (res *http.Response, err error) {
	if contentType == "" {
		contentType = appJson
	}

	req := Request{"POST", url, contentType, body, token}
	res, err = Do(client, &req)
	return
}

func Delete(client *http.Client, url string, token string) (res *http.Response, err error) {
	req := Request{"DELETE", url, "", nil, token}
	res, err = Do(client, &req)
	return
}

func Do(client *http.Client, req *Request) (res *http.Response, err error) {
	r, err := http.NewRequest(req.Method, req.URL, req.Body)
	if err != nil {
		return
	}
	if req.ContentType != "" {
		r.Header.Add("Content-Type", req.ContentType)
	}
	if req.Token != "" {
		r.Header.Add("Authorization", "Bearer "+req.Token)
	}
	res, err = client.Do(r)
	return
}

func MultipartUploadFile(client *http.Client, url, filePath string, params map[string]string, token string) (res *http.Response, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	return MultipartUpload(client, url, file, filepath.Base(filePath), params, token)
}

func MultipartUpload(client *http.Client, url string, reader io.Reader, filename string, params map[string]string, token string) (res *http.Response, err error) {
	// The mime/multipart package does not support streaming multipart data from disk,
	// at least not without complicated, problematic goroutines that simultaneously read/write into a buffer.
	// A much easier approach is to just construct the multipart request by hand, using io.MultiPart to
	// concatenate the parts of the request together into a single io.Reader.
	boundary := randomBoundary()
	parts := []io.Reader{}

	// Create a part for each key, val pair in params
	for k, v := range params {
		parts = append(parts, createFieldPart(k, v, boundary))
	}

	start := fmt.Sprintf("\r\n--%s\r\n", boundary)
	start += fmt.Sprintf("Content-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\n", quoteEscaper.Replace(filename))
	start += fmt.Sprintf("Content-Type: application/octet-stream\r\n\r\n")
	end := fmt.Sprintf("\r\n--%s--", boundary)

	// The request will consist of a reader to begin the request, a reader which points
	// to the file data on disk, and a reader containing the closing boundary of the request.
	parts = append(parts, strings.NewReader(start), reader, strings.NewReader(end))

	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)

	res, err = Do(client, &Request{"POST", url, contentType, io.MultiReader(parts...), token})

	return
}

// From https://golang.org/src/mime/multipart/writer.go
func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

// From https://golang.org/src/mime/multipart/writer.go
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// Creates a reader that encapsulates a single multipart form part
func createFieldPart(fieldname, value, boundary string) io.Reader {
	str := fmt.Sprintf("\r\n--%s\r\n", boundary)
	str += fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"\r\n\r\n", quoteEscaper.Replace(fieldname))
	str += value
	return strings.NewReader(str)
}
