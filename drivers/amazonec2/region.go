package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Ubuntu 16.04 LTS 20170619.1 hvm:ebs-ssd (amd64)
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1":  {"ami-785c491f"},
	"ap-northeast-2":  {"ami-94d20dfa"},
	"ap-southeast-1":  {"ami-2378f540"},
	"ap-southeast-2":  {"ami-e94e5e8a"},
	"ap-south-1":      {"ami-49e59a26"},
	"ca-central-1":    {"ami-7ed56a1a"},
	"cn-north-1":      {"ami-a163b4cc"}, // Note: this is 20170303
	"eu-central-1":    {"ami-1c45e273"},
	"eu-west-1":       {"ami-6d48500b"},
	"eu-west-2":       {"ami-cc7066a8"},
	"sa-east-1":       {"ami-34afc458"},
	"us-east-1":       {"ami-d15a75c7"},
	"us-east-2":       {"ami-8b92b4ee"},
	"us-west-1":       {"ami-73f7da13"},
	"us-west-2":       {"ami-835b4efa"},
	"us-gov-west-1":   {"ami-939412f2"},
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
