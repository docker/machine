package aliyunecs

import (
	"crypto/rand"
	"errors"

	"github.com/denverdino/aliyungo/ecs"
)

var (
	errInvalidRegion  = errors.New("invalid region specified")
	errNoVpcs         = errors.New("No VPCs found in region")
	errMachineFailure = errors.New("Machine failed to start")
	errNoIP           = errors.New("No IP Address associated with the instance")
	errComplete       = errors.New("Complete")
)

const defaultUbuntuImageID = "ubuntu1404_64_20G_aliaegis_20150325.vhd"
const defaultUbuntuImagePrefix = "ubuntu1404_64_20G_"

func validateECSRegion(region string) (ecs.Region, error) {
	for _, v := range ecs.ValidRegions {
		if v == ecs.Region(region) {
			return v, nil
		}
	}

	return "", errInvalidRegion
}

const dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const paswordLen = 16

func randomPassword() string {
	var bytes = make([]byte, paswordLen)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}
