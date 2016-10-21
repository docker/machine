package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Release 16.04 LTS 20161013
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1":  {"ami-31892c50"},
	"ap-northeast-2":  {"ami-a3915acd"},
	"ap-southeast-1":  {"ami-18e7417b"},
	"ap-southeast-2":  {"ami-7be4d618"},
	"ap-south-1":      {"ami-7e94fe11"},
	"cn-north-1":      {"ami-d7c511ba"},
	"eu-central-1":    {"ami-597c8236"},
	"eu-west-1":       {"ami-c593deb6"},
	"sa-east-1":       {"ami-909b06fc"},
	"us-east-1":       {"ami-fd6e3bea"},
	"us-east-2":       {"ami-0a104a6f"},
	"us-west-1":       {"ami-73531b13"},
	"us-west-2":       {"ami-f1ca1091"},
	"us-gov-west-1":   {"ami-8df24aec"},
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
