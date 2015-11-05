package awsauth

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSignature2(t *testing.T) {
	// http://docs.aws.amazon.com/general/latest/gr/signature-version-2.html

	Convey("Given bogus credentials", t, func() {
		keys := *testCredV2

		// Mock time
		now = func() time.Time {
			parsed, _ := time.Parse(timeFormatV2, exampleReqTsV2)
			return parsed
		}

		Convey("Given a plain request that is unprepared", func() {
			request := test_plainRequestV2()

			Convey("The request should be prepared to be signed", func() {
				expectedUnsigned := test_unsignedRequestV2()
				prepareRequestV2(request, keys)
				So(request, ShouldResemble, expectedUnsigned)
			})
		})

		Convey("Given a prepared, but unsigned, request", func() {
			request := test_unsignedRequestV2()

			Convey("The canonical query string should be correct", func() {
				actual := canonicalQueryStringV2(request)
				expected := canonicalQsV2
				So(actual, ShouldEqual, expected)
			})

			Convey("The absolute path should be extracted correctly", func() {
				So(request.URL.Path, ShouldEqual, "/")
			})

			Convey("The string to sign should be well-formed", func() {
				actual := stringToSignV2(request)
				So(actual, ShouldEqual, expectedStringToSignV2)
			})

			Convey("The resulting signature should be correct", func() {
				actual := signatureV2(stringToSignV2(request), keys)
				So(actual, ShouldEqual, "i91nKc4PWAt0JJIdXwz9HxZCJDdiy6cf/Mj6vPxyYIs=")
			})

			Convey("The final signed request should be correctly formed", func() {
				Sign2(request, keys)
				actual := request.URL.String()
				So(actual, ShouldResemble, expectedFinalUrlV2)
			})
		})
	})
}

func TestVersion2STSRequestPreparer(t *testing.T) {
	Convey("Given a plain request ", t, func() {
		request := test_plainRequestV2()

		Convey("And a set of credentials with an STS token", func() {
			var keys Credentials
			keys = *testCredV2WithSTS

			Convey("It should include the SecurityToken parameter when the request is signed", func() {
				actualSigned := Sign2(request, keys)
				actual := actualSigned.URL.Query()["SecurityToken"][0]

				So(actual, ShouldNotBeBlank)
				So(actual, ShouldEqual, testCredV2WithSTS.SecurityToken)

			})
		})
	})

}

func test_plainRequestV2() *http.Request {
	values := url.Values{}
	values.Set("Action", "DescribeJobFlows")
	values.Set("Version", "2009-03-31")

	url := baseUrlV2 + "?" + values.Encode()

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	return request
}

func test_unsignedRequestV2() *http.Request {
	request := test_plainRequestV2()
	newUrl, _ := url.Parse(baseUrlV2 + "/?" + canonicalQsV2)
	request.URL = newUrl
	return request
}

var (
	testCredV2 = &Credentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	testCredV2WithSTS = &Credentials{
		AccessKeyID:     "AKIDEXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		SecurityToken:   "AQoDYXdzEHcaoAJ1Aqwx1Sum0iW2NQjXJcWlKR7vuB6lnAeGBaQnjDRZPVyniwc48ml5hx+0qiXenVJdfusMMl9XLhSncfhx9Rb1UF8IAOaQ+CkpWXvoH67YYN+93dgckSVgVEBRByTl/BvLOZhe0ii/pOWkuQtBm5T7lBHRe4Dfmxy9X6hd8L3FrWxgnGV3fWZ3j0gASdYXaa+VBJlU0E2/GmCzn3T+t2mjYaeoInAnYVKVpmVMOrh6lNAeETTOHElLopblSa7TAmROq5xHIyu4a9i2qwjERTwa3Yk4Jk6q7JYVA5Cu7kS8wKVml8LdzzCTsy+elJgvH+Jf6ivpaHt/En0AJ5PZUJDev2+Y5+9j4AYfrmXfm4L73DC1ZJFJrv+Yh+EXAMPLE=",
	}

	exampleReqTsV2         = "2011-10-03T15:19:30"
	baseUrlV2              = "https://elasticmapreduce.amazonaws.com"
	canonicalQsV2          = "AWSAccessKeyId=AKIAIOSFODNN7EXAMPLE&Action=DescribeJobFlows&SignatureMethod=HmacSHA256&SignatureVersion=2&Timestamp=2011-10-03T15%3A19%3A30&Version=2009-03-31"
	expectedStringToSignV2 = "GET\nelasticmapreduce.amazonaws.com\n/\n" + canonicalQsV2
	expectedFinalUrlV2     = baseUrlV2 + "/?AWSAccessKeyId=AKIAIOSFODNN7EXAMPLE&Action=DescribeJobFlows&Signature=i91nKc4PWAt0JJIdXwz9HxZCJDdiy6cf%2FMj6vPxyYIs%3D&SignatureMethod=HmacSHA256&SignatureVersion=2&Timestamp=2011-10-03T15%3A19%3A30&Version=2009-03-31"
)
