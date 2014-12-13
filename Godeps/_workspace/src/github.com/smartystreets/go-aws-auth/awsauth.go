// Package awsauth implements AWS request signing using Signed Signature Version 2,
// Signed Signature Version 3, and Signed Signature Version 4. Supports S3 and STS.
package awsauth

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Credentials stores the information necessary to authorize with AWS and it
// is from this information that requests are signed.
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SecurityToken   string `json:"Token"`
	Expiration      time.Time
}

// Sign signs a request bound for AWS. It automatically chooses the best
// authentication scheme based on the service the request is going to.
func Sign(req *http.Request, cred ...Credentials) *http.Request {
	service, _ := serviceAndRegion(req.URL.Host)
	sigVersion := awsSignVersion[service]

	switch sigVersion {
	case 2:
		return Sign2(req, cred...)
	case 3:
		return Sign3(req, cred...)
	case 4:
		return Sign4(req, cred...)
	case -1:
		return SignS3(req, cred...)
	}

	return nil
}

// Sign4 signs a request with Signed Signature Version 4.
func Sign4(req *http.Request, cred ...Credentials) *http.Request {
	signMutex.Lock()
	defer signMutex.Unlock()
	keys := chooseKeys(cred)

	// Add the X-Amz-Security-Token header when using STS
	if keys.SecurityToken != "" {
		req.Header.Set("X-Amz-Security-Token", keys.SecurityToken)
	}

	prepareRequestV4(req)
	meta := new(metadata)

	// Task 1
	hashedCanonReq := hashedCanonicalRequestV4(req, meta)

	// Task 2
	stringToSign := stringToSignV4(req, hashedCanonReq, meta)

	// Task 3
	signingKey := signingKeyV4(keys.SecretAccessKey, meta.date, meta.region, meta.service)
	signature := signatureV4(signingKey, stringToSign)

	req.Header.Set("Authorization", buildAuthHeaderV4(signature, meta, keys))

	return req
}

// Sign3 signs a request with Signed Signature Version 3.
// If the service you're accessing supports Version 4, use that instead.
func Sign3(req *http.Request, cred ...Credentials) *http.Request {
	signMutex.Lock()
	defer signMutex.Unlock()
	keys := chooseKeys(cred)

	// Add the X-Amz-Security-Token header when using STS
	if keys.SecurityToken != "" {
		req.Header.Set("X-Amz-Security-Token", keys.SecurityToken)
	}

	prepareRequestV3(req)

	// Task 1
	stringToSign := stringToSignV3(req)

	// Task 2
	signature := signatureV3(stringToSign, keys)

	// Task 3
	req.Header.Set("X-Amzn-Authorization", buildAuthHeaderV3(signature, keys))

	return req
}

// Sign2 signs a request with Signed Signature Version 2.
// If the service you're accessing supports Version 4, use that instead.
func Sign2(req *http.Request, cred ...Credentials) *http.Request {
	signMutex.Lock()
	defer signMutex.Unlock()
	keys := chooseKeys(cred)

	// Add the SecurityToken parameter when using STS
	// This must be added before the signature is calculated
	if keys.SecurityToken != "" {
		v := url.Values{}
		v.Set("SecurityToken", keys.SecurityToken)
		augmentRequestQuery(req, v)

	}

	prepareRequestV2(req, keys)

	stringToSign := stringToSignV2(req)
	signature := signatureV2(stringToSign, keys)

	values := url.Values{}
	values.Set("Signature", signature)

	augmentRequestQuery(req, values)

	return req
}

// SignS3 signs a request bound for Amazon S3 using their custom
// HTTP authentication scheme.
func SignS3(req *http.Request, cred ...Credentials) *http.Request {
	signMutex.Lock()
	defer signMutex.Unlock()
	keys := chooseKeys(cred)

	// Add the X-Amz-Security-Token header when using STS
	if keys.SecurityToken != "" {
		req.Header.Set("X-Amz-Security-Token", keys.SecurityToken)
	}

	prepareRequestS3(req)

	stringToSign := stringToSignS3(req)
	signature := signatureS3(stringToSign, keys)

	authHeader := "AWS " + keys.AccessKeyID + ":" + signature
	req.Header.Set("Authorization", authHeader)

	return req
}

// SignS3Url signs a GET request for a resource on Amazon S3 by appending
// query string parameters containing credentials and signature. You must
// specify an expiration date for these signed requests. After that date,
// a request signed with this method will be rejected by S3.
func SignS3Url(req *http.Request, expire time.Time, cred ...Credentials) *http.Request {
	signMutex.Lock()
	defer signMutex.Unlock()
	keys := chooseKeys(cred)

	stringToSign := stringToSignS3Url("GET", expire, req.URL.Path)
	signature := signatureS3(stringToSign, keys)

	qs := req.URL.Query()
	qs.Set("AWSAccessKeyId", keys.AccessKeyID)
	qs.Set("Signature", signature)
	qs.Set("Expires", timeToUnixEpochString(expire))
	req.URL.RawQuery = qs.Encode()

	return req
}

// expired checks to see if the temporary credentials from an IAM role are
// within 4 minutes of expiration (The IAM documentation says that new keys
// will be provisioned 5 minutes before the old keys expire). Credentials
// that do not have an Expiration cannot expire.
func (k *Credentials) expired() bool {
	if k.Expiration.IsZero() {
		// Credentials with no expiration can't expire
		return false
	}
	expireTime := k.Expiration.Add(-4 * time.Minute)
	// if t - 4 mins is before now, true
	if expireTime.Before(time.Now()) {
		return true
	} else {
		return false
	}
}

type metadata struct {
	algorithm       string
	credentialScope string
	signedHeaders   string
	date            string
	region          string
	service         string
}

const (
	envAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envSecurityToken   = "AWS_SECURITY_TOKEN"
)

var (
	awsSignVersion = map[string]int{
		"autoscaling":          4,
		"cloudfront":           4,
		"cloudformation":       4,
		"cloudsearch":          4,
		"monitoring":           4,
		"dynamodb":             4,
		"ec2":                  2,
		"elasticmapreduce":     4,
		"elastictranscoder":    4,
		"elasticache":          2,
		"glacier":              4,
		"kinesis":              4,
		"redshift":             4,
		"rds":                  4,
		"sdb":                  2,
		"sns":                  4,
		"sqs":                  4,
		"s3":                   4,
		"elasticbeanstalk":     4,
		"importexport":         2,
		"iam":                  4,
		"route53":              3,
		"elasticloadbalancing": 4,
		"email":                3,
	}

	signMutex sync.Mutex
)
