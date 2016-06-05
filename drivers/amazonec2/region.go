package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Release 16.04 LTS 20160516.1
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1": {"ami-5d38d93c"},
	"ap-northeast-2": {"ami-a3915acd"},
	"ap-southeast-1": {"ami-a35284c0"},
	"ap-southeast-2": {"ami-f4361997"},
	"cn-north-1":     {"ami-79eb2214"}, // 15.10 20151116.1
	"eu-west-1":      {"ami-7a138709"},
	"eu-central-1":   {"ami-f9e30f96"},
	"sa-east-1":      {"ami-0d5dd561"},
	"us-east-1":      {"ami-13be557e"},
	"us-west-1":      {"ami-84423ae4"},
	"us-west-2":      {"ami-06b94666"},
	"us-gov-west-1":  {"ami-8f4df2ee"},
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
