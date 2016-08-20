package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Release 16.04 LTS 20160815
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1":  {"ami-51f13330"},
	"ap-northeast-2":  {"ami-a3915acd"}, // 20160516.1
	"ap-southeast-1":  {"ami-fec51c9d"},
	"ap-southeast-2":  {"ami-a78ebac4"},
	"ap-south-1":      {"ami-7e94fe11"}, // 20160627
	"cn-north-1":      {"ami-2c3bf141"}, // 20160610
	"eu-central-1":    {"ami-004abc6f"},
	"eu-west-1":       {"ami-c06b1eb3"},
	"sa-east-1":       {"ami-a674e2ca"},
	"us-east-1":       {"ami-c60b90d1"},
	"us-west-1":       {"ami-1bf0b37b"},
	"us-west-2":       {"ami-f701cb97"},
	"us-gov-west-1":   {"ami-76f34a17"},
	"custom-endpoint": {""},
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
