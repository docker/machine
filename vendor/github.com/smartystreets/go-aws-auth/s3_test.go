package awsauth

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignatureS3(t *testing.T) {
	// http://docs.aws.amazon.com/AmazonS3/2006-03-01/dev/RESTAuthentication.html
	// Note: S3 now supports signed signature version 4
	// (but signed URL requests still utilize a lot of the same functionality)

	Convey("Given a GET request to Amazon S3", t, func() {
		keys := *testCredS3
		req := test_plainRequestS3()

		// Mock time
		now = func() time.Time {
			parsed, _ := time.Parse(timeFormatS3, exampleReqTsS3)
			return parsed
		}

		Convey("The request should be prepared with a Date header", func() {
			prepareRequestS3(req)
			So(req.Header.Get("Date"), ShouldEqual, exampleReqTsS3)
		})

		Convey("The CanonicalizedAmzHeaders should be built properly", func() {
			req2 := test_headerRequestS3()
			actual := canonicalAmzHeadersS3(req2)
			So(actual, ShouldEqual, expectedCanonAmzHeadersS3)
		})

		Convey("The CanonicalizedResource should be built properly", func() {
			actual := canonicalResourceS3(req)
			So(actual, ShouldEqual, expectedCanonResourceS3)
		})

		Convey("The string to sign should be correct", func() {
			actual := stringToSignS3(req)
			So(actual, ShouldEqual, expectedStringToSignS3)
		})

		Convey("The final signature string should be exactly correct", func() {
			actual := signatureS3(stringToSignS3(req), keys)
			So(actual, ShouldEqual, "bWq2s1WEIj+Ydj0vQ697zp+IXMU=")
		})
	})

	Convey("Given a GET request for a resource on S3 for query string authentication", t, func() {
		keys := *testCredS3
		req, _ := http.NewRequest("GET", "https://johnsmith.s3.amazonaws.com/johnsmith/photos/puppy.jpg", nil)

		now = func() time.Time {
			parsed, _ := time.Parse(timeFormatS3, exampleReqTsS3)
			return parsed
		}

		Convey("The string to sign should be correct", func() {
			actual := stringToSignS3Url("GET", now(), req.URL.Path)
			So(actual, ShouldEqual, expectedStringToSignS3Url)
		})

		Convey("The signature of string to sign should be correct", func() {
			actual := signatureS3(expectedStringToSignS3Url, keys)
			So(actual, ShouldEqual, "R2K/+9bbnBIbVDCs7dqlz3XFtBQ=")
		})

		Convey("The finished signed URL should be correct", func() {
			expiry := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
			So(SignS3Url(req, expiry, keys).URL.String(), ShouldEqual, expectedSignedS3Url)
		})
	})
}

func TestS3STSRequestPreparer(t *testing.T) {
	Convey("Given a plain request with no custom headers", t, func() {
		req := test_plainRequestS3()

		Convey("And a set of credentials with an STS token", func() {
			keys := *testCredS3WithSTS

			Convey("It should include an X-Amz-Security-Token when the request is signed", func() {
				actualSigned := SignS3(req, keys)
				actual := actualSigned.Header.Get("X-Amz-Security-Token")

				So(actual, ShouldNotBeBlank)
				So(actual, ShouldEqual, testCredS3WithSTS.SecurityToken)

			})
		})
	})
}

func test_plainRequestS3() *http.Request {
	req, _ := http.NewRequest("GET", "https://johnsmith.s3.amazonaws.com/photos/puppy.jpg", nil)
	return req
}

func test_headerRequestS3() *http.Request {
	req := test_plainRequestS3()
	req.Header.Set("X-Amz-Meta-Something", "more foobar")
	req.Header.Set("X-Amz-Date", "foobar")
	req.Header.Set("X-Foobar", "nanoo-nanoo")
	return req
}

func TestCanonical(t *testing.T) {
	expectedCanonicalString := "PUT\nc8fdb181845a4ca6b8fec737b3581d76\ntext/html\nThu, 17 Nov 2005 18:49:58 GMT\nx-amz-magic:abracadabra\nx-amz-meta-author:foo@bar.com\n/quotes/nelson"

	origUrl := "https://s3.amazonaws.com/"
	resource := "/quotes/nelson"

	u, _ := url.ParseRequestURI(origUrl)
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	req, _ := http.NewRequest("PUT", urlStr, nil)
	req.Header.Add("Content-Md5", "c8fdb181845a4ca6b8fec737b3581d76")
	req.Header.Add("Content-Type", "text/html")
	req.Header.Add("Date", "Thu, 17 Nov 2005 18:49:58 GMT")
	req.Header.Add("X-Amz-Meta-Author", "foo@bar.com")
	req.Header.Add("X-Amz-Magic", "abracadabra")

	if stringToSignS3(req) != expectedCanonicalString {
		t.Errorf("----Got\n***%s***\n----Expected\n***%s***", stringToSignS3(req), expectedCanonicalString)
	}
}

var (
	testCredS3 = &Credentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	testCredS3WithSTS = &Credentials{
		AccessKeyID:     "AKIDEXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		SecurityToken:   "AQoDYXdzEHcaoAJ1Aqwx1Sum0iW2NQjXJcWlKR7vuB6lnAeGBaQnjDRZPVyniwc48ml5hx+0qiXenVJdfusMMl9XLhSncfhx9Rb1UF8IAOaQ+CkpWXvoH67YYN+93dgckSVgVEBRByTl/BvLOZhe0ii/pOWkuQtBm5T7lBHRe4Dfmxy9X6hd8L3FrWxgnGV3fWZ3j0gASdYXaa+VBJlU0E2/GmCzn3T+t2mjYaeoInAnYVKVpmVMOrh6lNAeETTOHElLopblSa7TAmROq5xHIyu4a9i2qwjERTwa3Yk4Jk6q7JYVA5Cu7kS8wKVml8LdzzCTsy+elJgvH+Jf6ivpaHt/En0AJ5PZUJDev2+Y5+9j4AYfrmXfm4L73DC1ZJFJrv+Yh+EXAMPLE=",
	}

	expectedCanonAmzHeadersS3 = "x-amz-date:foobar\nx-amz-meta-something:more foobar\n"
	expectedCanonResourceS3   = "/johnsmith/photos/puppy.jpg"
	expectedStringToSignS3    = "GET\n\n\nTue, 27 Mar 2007 19:36:42 +0000\n/johnsmith/photos/puppy.jpg"
	expectedStringToSignS3Url = "GET\n\n\n1175024202\n/johnsmith/photos/puppy.jpg"
	expectedSignedS3Url       = "https://johnsmith.s3.amazonaws.com/johnsmith/photos/puppy.jpg?AWSAccessKeyId=AKIAIOSFODNN7EXAMPLE&Expires=1257894000&Signature=X%2FarTLAJP08uP1Bsap52rwmsVok%3D"
	exampleReqTsS3            = "Tue, 27 Mar 2007 19:36:42 +0000"
)
