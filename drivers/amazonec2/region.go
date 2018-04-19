package amazonec2

import (
	"errors"
)

type region struct {
	AmiId              string
	InstanceStoreAmiId string
}

// Ubuntu 16.04 LTS 20180405 hvm:ebs-ssd + hvm:instance-store (amd64)
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1":  {"ami-60a4b21c", "ami-55564729"},
	"ap-northeast-2":  {"ami-633d920d", "ami-c63e91a8"},
	"ap-northeast-3":  {"ami-89c3cdf4", "ami-b4c2ccc9"},
	"ap-southeast-1":  {"ami-82c9ecfe", "ami-c4d7f2b8"},
	"ap-southeast-2":  {"ami-2b12dc49", "ami-7e2ee01c"},
	"ap-south-1":      {"ami-dba580b4", "ami-bba287d4"},
	"ca-central-1":    {"ami-9d7afcf9", "ami-24048240"},
	"cn-north-1":      {"ami-a209d6cf", "ami-4f06d922"},
	"cn-northwest-1":  {"ami-fd0e1a9f", ""}, // Note: this is 20180126
	"eu-central-1":    {"ami-cd491726", "ami-59421cb2"},
	"eu-west-1":       {"ami-74e6b80d", "ami-85e5bbfc"},
	"eu-west-2":       {"ami-506e8f37", "ami-7e739219"},
	"eu-west-3":       {"ami-9a03b5e7", "ami-9b0fb9e6"},
	"sa-east-1":       {"ami-5782d43b", "ami-9385d3ff"},
	"us-east-1":       {"ami-6dfe5010", "ami-bbcf61c6"},
	"us-east-2":       {"ami-e82a1a8d", "ami-d82a1abd"},
	"us-west-1":       {"ami-493f2f29", "ami-4b21312b"},
	"us-west-2":       {"ami-ca89eeb2", "ami-2c95f254"},
	"us-gov-west-1":   {"ami-fb77e29a", "ami-7f70e51e"},
	"custom-endpoint": {"", ""},
}

func awsRegionsList() []string {
	var list []string

	for k := range regionDetails {
		list = append(list, k)
	}

	return list
}

func validateAwsRegion(region string) (string, error) {
	for _, v := range awsRegionsList() {
		if v == region {
			return region, nil
		}
	}

	return "", errors.New("Invalid region specified")
}
