package awsauth

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSignature3(t *testing.T) {
	// http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/RESTAuthentication.html
	// http://docs.aws.amazon.com/ses/latest/DeveloperGuide/query-interface-authentication.html

	Convey("Given bogus credentials", t, func() {
		keys := *testCredV3

		// Mock time
		now = func() time.Time {
			parsed, _ := time.Parse(timeFormatV3, exampleReqTsV3)
			return parsed
		}

		Convey("Given a plain request that is unprepared", func() {
			request := test_plainRequestV3()

			Convey("The request should be prepared to be signed", func() {
				expectedUnsigned := test_unsignedRequestV3()
				prepareRequestV3(request)
				So(request, ShouldResemble, expectedUnsigned)
			})
		})

		Convey("Given a prepared, but unsigned, request", func() {
			request := test_unsignedRequestV3()

			Convey("The absolute path should be extracted correctly", func() {
				So(request.URL.Path, ShouldEqual, "/")
			})

			Convey("The string to sign should be well-formed", func() {
				actual := stringToSignV3(request)
				So(actual, ShouldEqual, expectedStringToSignV3)
			})

			Convey("The resulting signature should be correct", func() {
				actual := signatureV3(stringToSignV3(request), keys)
				So(actual, ShouldEqual, "PjAJ6buiV6l4WyzmmuwtKE59NJXVg5Dr3Sn4PCMZ0Yk=")
			})

			Convey("The final signed request should be correctly formed", func() {
				Sign3(request, keys)
				actual := request.Header.Get("X-Amzn-Authorization")
				So(actual, ShouldResemble, expectedAuthHeaderV3)
			})
		})
	})
}

func test_plainRequestV3() *http.Request {
	values := url.Values{}
	values.Set("Action", "GetSendStatistics")
	values.Set("Version", "2010-12-01")

	url := baseUrlV3 + "/?" + values.Encode()

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	return request
}

func test_unsignedRequestV3() *http.Request {
	request := test_plainRequestV3()
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	request.Header.Set("x-amz-date", exampleReqTsV3)
	request.Header.Set("Date", exampleReqTsV3)
	request.Header.Set("x-amz-nonce", "")
	return request
}

func TestVersion3STSRequestPreparer(t *testing.T) {
	Convey("Given a plain request with no custom headers", t, func() {
		request := test_plainRequestV3()

		Convey("And a set of credentials with an STS token", func() {
			var keys Credentials
			keys = *testCredV3WithSTS

			Convey("It should include an X-Amz-Security-Token when the request is signed", func() {
				actualSigned := Sign3(request, keys)
				actual := actualSigned.Header.Get("X-Amz-Security-Token")

				So(actual, ShouldNotBeBlank)
				So(actual, ShouldEqual, testCredV4WithSTS.SecurityToken)

			})
		})
	})

}

var (
	testCredV3 = &Credentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	testCredV3WithSTS = &Credentials{
		AccessKeyID:     "AKIDEXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		SecurityToken:   "AQoDYXdzEHcaoAJ1Aqwx1Sum0iW2NQjXJcWlKR7vuB6lnAeGBaQnjDRZPVyniwc48ml5hx+0qiXenVJdfusMMl9XLhSncfhx9Rb1UF8IAOaQ+CkpWXvoH67YYN+93dgckSVgVEBRByTl/BvLOZhe0ii/pOWkuQtBm5T7lBHRe4Dfmxy9X6hd8L3FrWxgnGV3fWZ3j0gASdYXaa+VBJlU0E2/GmCzn3T+t2mjYaeoInAnYVKVpmVMOrh6lNAeETTOHElLopblSa7TAmROq5xHIyu4a9i2qwjERTwa3Yk4Jk6q7JYVA5Cu7kS8wKVml8LdzzCTsy+elJgvH+Jf6ivpaHt/En0AJ5PZUJDev2+Y5+9j4AYfrmXfm4L73DC1ZJFJrv+Yh+EXAMPLE=",
	}

	exampleReqTsV3         = "Thu, 14 Aug 2008 17:08:48 GMT"
	baseUrlV3              = "https://email.us-east-1.amazonaws.com"
	expectedStringToSignV3 = exampleReqTsV3
	expectedAuthHeaderV3   = "AWS3-HTTPS AWSAccessKeyId=" + testCredV3.AccessKeyID + ", Algorithm=HmacSHA256, Signature=PjAJ6buiV6l4WyzmmuwtKE59NJXVg5Dr3Sn4PCMZ0Yk="
)
