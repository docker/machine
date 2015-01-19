package awsec2

import (
	"errors"
)

var (
	errInvalidRegion = errors.New("invalid region specified")
	errNoVpcs        = errors.New("No VPCs found in region")
	errComplete      = errors.New("Complete")
)

type region struct {
	AmiId string
}

var regionDetails map[string]*region = map[string]*region{
	"ap-northeast-1": &region{"ami-44f1e245"},
	"ap-southeast-1": &region{"ami-f95875ab"},
	"ap-southeast-2": &region{"ami-890b62b3"},
	"cn-north-1":     &region{"ami-fe7ae8c7"},
	"eu-west-1":      &region{"ami-823686f5"},
	"eu-central-1":   &region{"ami-ac1524b1"},
	"sa-east-1":      &region{"ami-c770c1da"},
	"us-east-1":      &region{"ami-4ae27e22"},
	"us-west-1":      &region{"ami-d1180894"},
	"us-west-2":      &region{"ami-898dd9b9"},
	"us-gov-west-1":  &region{"ami-cf5630ec"},
}

func awsRegionsList() []string {
	var list []string

	for k, _ := range regionDetails {
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
