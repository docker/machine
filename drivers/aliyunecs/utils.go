package aliyunecs

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
	ImageID string
}

//TODO Match the latest Ubuntu 1404 image automatically
const IMAGE_ID = "ubuntu1404_64_20G_aliaegis_20150325.vhd"

var regionDetails map[string]*region = map[string]*region{
	"cn-shenzhen": {IMAGE_ID},
	"cn-qingdao":  {IMAGE_ID},
	"cn-beijing":  {IMAGE_ID},
	"cn-hongkong": {IMAGE_ID},
	"cn-hangzhou": {IMAGE_ID},
	"us-west-1":   {IMAGE_ID},
}

func ecsRegionsList() []string {
	var list []string

	for k := range regionDetails {
		list = append(list, k)
	}

	return list
}

func validateECSRegion(region string) (string, error) {
	for _, v := range ecsRegionsList() {
		if v == region {
			return region, nil
		}
	}

	return "", errInvalidRegion
}
