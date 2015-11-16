package awsauth

import (
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func signatureS3(stringToSign string, keys Credentials) string {
	hashed := hmacSHA1([]byte(keys.SecretAccessKey), stringToSign)
	return base64.StdEncoding.EncodeToString(hashed)
}

func stringToSignS3(req *http.Request) string {
	str := req.Method + "\n"

	if req.Header.Get("Content-Md5") != "" {
		str += req.Header.Get("Content-Md5")
	} else {
		body := readAndReplaceBody(req)
		if len(body) > 0 {
			str += hashMD5(body)
		}
	}
	str += "\n"

	str += req.Header.Get("Content-Type") + "\n"

	if req.Header.Get("Date") != "" {
		str += req.Header.Get("Date")
	} else {
		str += timestampS3()
	}

	str += "\n"

	canonicalHeaders := canonicalAmzHeadersS3(req)
	if canonicalHeaders != "" {
		str += canonicalHeaders
	}

	str += canonicalResourceS3(req)

	return str
}

func stringToSignS3Url(method string, expire time.Time, path string) string {
	return method + "\n\n\n" + timeToUnixEpochString(expire) + "\n" + path
}

func timeToUnixEpochString(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

func canonicalAmzHeadersS3(req *http.Request) string {
	var headers []string

	for header := range req.Header {
		standardized := strings.ToLower(strings.TrimSpace(header))
		if strings.HasPrefix(standardized, "x-amz") {
			headers = append(headers, standardized)
		}
	}

	sort.Strings(headers)

	for i, header := range headers {
		headers[i] = header + ":" + strings.Replace(req.Header.Get(header), "\n", " ", -1)
	}

	if len(headers) > 0 {
		return strings.Join(headers, "\n") + "\n"
	} else {
		return ""
	}
}

func canonicalResourceS3(req *http.Request) string {
	res := ""

	if isS3VirtualHostedStyle(req) {
		bucketname := strings.Split(req.Host, ".")[0]
		res += "/" + bucketname
	}

	res += req.URL.Path

	for _, subres := range strings.Split(subresourcesS3, ",") {
		if strings.HasPrefix(req.URL.RawQuery, subres) {
			res += "?" + subres
		}
	}

	return res
}

func prepareRequestS3(req *http.Request) *http.Request {
	req.Header.Set("Date", timestampS3())
	if req.URL.Path == "" {
		req.URL.Path += "/"
	}
	return req
}

// Info: http://docs.aws.amazon.com/AmazonS3/latest/dev/VirtualHosting.html
func isS3VirtualHostedStyle(req *http.Request) bool {
	service, _ := serviceAndRegion(req.Host)
	return service == "s3" && strings.Count(req.Host, ".") == 3
}

func timestampS3() string {
	return now().Format(timeFormatS3)
}

const (
	timeFormatS3   = time.RFC1123Z
	subresourcesS3 = "acl,lifecycle,location,logging,notification,partNumber,policy,requestPayment,torrent,uploadId,uploads,versionId,versioning,versions,website"
)
