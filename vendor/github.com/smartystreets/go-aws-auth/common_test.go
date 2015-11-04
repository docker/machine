package awsauth

import (
	"testing"
	"net/url"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCommonFunctions(t *testing.T) {
	Convey("Service and region should be properly extracted from host strings", t, func() {
		service, region := serviceAndRegion("sqs.us-west-2.amazonaws.com")
		So(service, ShouldEqual, "sqs")
		So(region, ShouldEqual, "us-west-2")

		service, region = serviceAndRegion("iam.amazonaws.com")
		So(service, ShouldEqual, "iam")
		So(region, ShouldEqual, "us-east-1")

		service, region = serviceAndRegion("sns.us-west-2.amazonaws.com")
		So(service, ShouldEqual, "sns")
		So(region, ShouldEqual, "us-west-2")

		service, region = serviceAndRegion("bucketname.s3.amazonaws.com")
		So(service, ShouldEqual, "s3")
		So(region, ShouldEqual, "us-east-1")

		service, region = serviceAndRegion("s3.amazonaws.com")
		So(service, ShouldEqual, "s3")
		So(region, ShouldEqual, "us-east-1")

		service, region = serviceAndRegion("s3-us-west-1.amazonaws.com")
		So(service, ShouldEqual, "s3")
		So(region, ShouldEqual, "us-west-1")

		service, region = serviceAndRegion("s3-external-1.amazonaws.com")
		So(service, ShouldEqual, "s3")
		So(region, ShouldEqual, "us-east-1")
	})

	Convey("MD5 hashes should be properly computed and base-64 encoded", t, func() {
		input := []byte("Pretend this is a REALLY long byte array...")
		actual := hashMD5(input)

		So(actual, ShouldEqual, "KbVTY8Vl6VccnzQf1AGOFw==")
	})

	Convey("SHA-256 hashes should be properly hex-encoded (base 16)", t, func() {
		input := []byte("This is... Sparta!!")
		actual := hashSHA256(input)

		So(actual, ShouldEqual, "5c81a4ef1172e89b1a9d575f4cd82f4ed20ea9137e61aa7f1ab936291d24e79a")
	})

	Convey("Given a key and contents", t, func() {
		key := []byte("asdf1234")
		contents := "SmartyStreets was here"

		Convey("HMAC-SHA256 should be properly computed", func() {
			expected := []byte{65, 46, 186, 78, 2, 155, 71, 104, 49, 37, 5, 66, 195, 129, 159, 227, 239, 53, 240, 107, 83, 21, 235, 198, 238, 216, 108, 149, 143, 222, 144, 94}
			actual := hmacSHA256(key, contents)

			So(actual, ShouldResemble, expected)
		})

		Convey("HMAC-SHA1 should be properly computed", func() {
			expected := []byte{164, 77, 252, 0, 87, 109, 207, 110, 163, 75, 228, 122, 83, 255, 233, 237, 125, 206, 85, 70}
			actual := hmacSHA1(key, contents)

			So(actual, ShouldResemble, expected)
		})
	})

	Convey("Strings should be properly concatenated with a delimiter", t, func() {
		So(concat("\n", "Test1", "Test2"), ShouldEqual, "Test1\nTest2")
		So(concat(".", "Test1"), ShouldEqual, "Test1")
		So(concat("\t", "1", "2", "3", "4"), ShouldEqual, "1\t2\t3\t4")
	})

	Convey("URI components should be properly encoded", t, func() {
		So(normuri("/-._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"), ShouldEqual, "/-._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
		So(normuri("/ /foo"), ShouldEqual, "/%20/foo")
		So(normuri("/(foo)"), ShouldEqual, "/%28foo%29")
	})

	Convey("URI query strings should be properly encoded", t, func() {
		So(normquery(url.Values{"p": []string{" +&;-=._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"}}), ShouldEqual, "p=%20%2B%26%3B-%3D._~0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	})
}
