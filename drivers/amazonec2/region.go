package amazonec2

import (
	"errors"
)

type region struct {
	AmiId string
}

// Ubuntu 20.04 LTS 20210223 hvm:ebs-ssd (amd64)
// See https://cloud-images.ubuntu.com/locator/ec2/
var regionDetails map[string]*region = map[string]*region{
	"af-south-1":      {"ami-0081edcfb10f9f0d6"},
	"ap-east-1":       {"ami-0774445f9e6290ccd"},
	"ap-northeast-1":  {"ami-059b6d3840b03d6dd"},
	"ap-northeast-2":  {"ami-00f1068284b9eca92"},
	"ap-northeast-3":  {"ami-01ecbd21b1e9b987f"},
	"ap-southeast-1":  {"ami-01581ffba3821cdf3"},
	"ap-southeast-2":  {"ami-0a43280cfb87ffdba"},
	"ap-south-1":      {"ami-0d758c1134823146a"},
	"ca-central-1":    {"ami-043e33039f1a50a56"},
	"cn-north-1":      {"ami-0592ccadb56e65f8d"}, // Note: this is 20201112.1
	"cn-northwest-1":  {"ami-007d0f254ea0f8588"}, // Note: this is 20201112.1
	"eu-north-1":      {"ami-0ed17ff3d78e74700"},
	"eu-south-1":      {"ami-08a7e27b95390cc06"},
	"eu-central-1":    {"ami-0767046d1677be5a0"},
	"eu-west-1":       {"ami-08bac620dc84221eb"},
	"eu-west-2":       {"ami-096cb92bb3580c759"},
	"eu-west-3":       {"ami-0d6aecf0f0425f42a"},
	"me-south-1":      {"ami-07d42d0c2a45aa449"},
	"sa-east-1":       {"ami-0b9517e2052e8be7a"},
	"us-east-1":       {"ami-042e8287309f5df03"},
	"us-east-2":       {"ami-08962a4068733a2b6"},
	"us-west-1":       {"ami-031b673f443c2172c"},
	"us-west-2":       {"ami-0ca5c3bd5a268e7db"},
	"us-gov-west-1":   {"ami-a7edd7c6"}, // Note: this is 20210119.1
	"us-gov-east-1":   {"ami-c39973b2"}, // Note: this is 20210119.1
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
