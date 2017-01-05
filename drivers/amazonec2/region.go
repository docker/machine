package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Ubuntu 16.04 LTS 20161221 hvm:ebs-ssd (amd64)
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1":  {"ami-18afc47f"},
	"ap-northeast-2":  {"ami-93d600fd"},
	"ap-southeast-1":  {"ami-87b917e4"},
	"ap-southeast-2":  {"ami-e6b58e85"},
	"ap-south-1":      {"ami-dd3442b2"},
	"ca-central-1":    {"ami-7112a015"},
	"cn-north-1":      {"ami-31499d5c"},
	"eu-central-1":    {"ami-fe408091"},
	"eu-west-1":       {"ami-ca80a0b9"},
	"eu-west-2":       {"ami-ede2e889"},
	"sa-east-1":       {"ami-e075ed8c"},
	"us-east-1":       {"ami-9dcfdb8a"},
	"us-east-2":       {"ami-fcc19b99"},
	"us-west-1":       {"ami-b05203d0"},
	"us-west-2":       {"ami-b2d463d2"},
	"us-gov-west-1":   {"ami-19d56d78"},
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
