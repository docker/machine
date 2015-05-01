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

var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1": {"ami-fc11d4fc"},
	"ap-southeast-1": {"ami-7854692a"},
	"ap-southeast-2": {"ami-c5611cff"},
	"cn-north-1":     {"ami-7cd84545"},
	"eu-west-1":      {"ami-2d96f65a"},
	"eu-central-1":   {"ami-3cdae621"},
	"sa-east-1":      {"ami-71b2376c"},
	"us-east-1":      {"ami-cc3b3ea4"},
	"us-west-1":      {"ami-017f9d45"},
	"us-west-2":      {"ami-55526765"},
	"us-gov-west-1":  {"ami-8ffa9bac"},
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
