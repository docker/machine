package awsauth

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
)

func prepareRequestV2(req *http.Request, keys Credentials) *http.Request {

	keyID := keys.AccessKeyID

	values := url.Values{}
	values.Set("AWSAccessKeyId", keyID)
	values.Set("SignatureVersion", "2")
	values.Set("SignatureMethod", "HmacSHA256")
	values.Set("Timestamp", timestampV2())

	augmentRequestQuery(req, values)

	if req.URL.Path == "" {
		req.URL.Path += "/"
	}

	return req
}

func stringToSignV2(req *http.Request) string {
	str := req.Method + "\n"
	str += strings.ToLower(req.URL.Host) + "\n"
	str += req.URL.Path + "\n"
	str += canonicalQueryStringV2(req)
	return str
}

func signatureV2(strToSign string, keys Credentials) string {
	hashed := hmacSHA256([]byte(keys.SecretAccessKey), strToSign)
	return base64.StdEncoding.EncodeToString(hashed)
}

func canonicalQueryStringV2(req *http.Request) string {
	return req.URL.RawQuery
}

func timestampV2() string {
	return now().Format(timeFormatV2)
}

const timeFormatV2 = "2006-01-02T15:04:05"
