package amazonec2

import (
	"errors"
)

var (
	errInvalidRegion  = errors.New("invalid region specified")
	errNoVpcs         = errors.New("No VPCs found in region")
	errMachineFailure = errors.New("Machine failed to start")
	errNoIP           = errors.New("No IP Address associated with the instance")
	errComplete       = errors.New("Complete")
)

type region struct {
	AmiId string
}

// Release 20150603
var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1": {"ami-f4b06cf4"},
	"ap-southeast-1": {"ami-b899a2ea"},
	"ap-southeast-2": {"ami-b59ce48f"},
	"cn-north-1":     {"ami-da930ee3"},
	"eu-west-1":      {"ami-45d8a532"},
	"eu-central-1":   {"ami-b6e0d9ab"},
	"sa-east-1":      {"ami-1199190c"},
	"us-east-1":      {"ami-5f709f34"},
	"us-west-1":      {"ami-615cb725"},
	"us-west-2":      {"ami-7f675e4f"},
	"us-gov-west-1":  {"ami-99a9c9ba"},
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

	return "", errInvalidRegion
}
