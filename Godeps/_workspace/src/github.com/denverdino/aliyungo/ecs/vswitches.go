package ecs

import (
	"github.com/denverdino/aliyungo/util"
	"time"
)

type CreateVSwitchArgs struct {
	ZoneId      string
	CidrBlock   string
	VpcId       string
	VSwitchName string
	Description string
	ClientToken string
}

type CreateVSwitchResponse struct {
	CommonResponse
	VSwitchId string
}

// CreateVSwitch creates Virtual Switch
func (client *Client) CreateVSwitch(args *CreateVSwitchArgs) (vswitchId string, err error) {
	response := CreateVSwitchResponse{}
	err = client.Invoke("CreateVSwitch", args, &response)
	if err != nil {
		return "", err
	}
	return response.VSwitchId, err
}

type DeleteVSwitchArgs struct {
	VSwitchId string
}

type DeleteVSwitchResponse struct {
	CommonResponse
}

// DeleteVSwitch deletes Virtual Switch
func (client *Client) DeleteVSwitch(VSwitchId string) error {
	args := DeleteVSwitchArgs{
		VSwitchId: VSwitchId,
	}
	response := DeleteVSwitchResponse{}
	return client.Invoke("DeleteVSwitch", &args, &response)
}

type DescribeVSwitchesArgs struct {
	VpcId     string
	VSwitchId string
	ZoneId    string
	RegionId  Region
	Pagination
}

type VSwitchStatus string

const (
	VSwitchStatusPending   = VSwitchStatus("Pending")
	VSwitchStatusAvailable = VSwitchStatus("Available")
)

type VSwitchSetType struct {
	VSwitchId               string
	VpcId                   string
	Status                  VSwitchStatus // enum Pending | Available
	CidrBlock               string
	ZoneId                  string
	AvailableIpAddressCount int
	Description             string
	VSwitchName             string
	CreationTime            util.ISO6801Time
}

type DescribeVSwitchesResponse struct {
	CommonResponse
	PaginationResult
	VSwitches struct {
		VSwitch []VSwitchSetType
	}
}

// DescribeVSwitches describes Virtual Switches
func (client *Client) DescribeVSwitches(args *DescribeVSwitchesArgs) (vswitches []VSwitchSetType, pagination *PaginationResult, err error) {
	args.validate()
	response := DescribeVSwitchesResponse{}

	err = client.Invoke("DescribeVSwitches", args, &response)

	if err == nil {
		return response.VSwitches.VSwitch, &response.PaginationResult, nil
	}

	return nil, nil, err
}

type ModifyVSwitchAttributeArgs struct {
	VSwitchId   string
	VSwitchName string
	Description string
}

type ModifyVSwitchAttributeResponse struct {
	CommonResponse
}

// ModifyVSwitchAttribute modifies attribute of Virtual Private Cloud
func (client *Client) ModifyVSwitchAttribute(args *ModifyVSwitchAttributeArgs) error {
	response := ModifyVSwitchAttributeResponse{}
	return client.Invoke("ModifyVSwitchAttribute", args, &response)
}

// WaitForVSwitchAvailable waits for VSwitch to given status
func (client *Client) WaitForVSwitchAvailable(vpcId string, vswitchId string, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	args := DescribeVSwitchesArgs{
		VpcId:     vpcId,
		VSwitchId: vswitchId,
	}
	for {
		vpcs, _, err := client.DescribeVSwitches(&args)
		if err != nil {
			return err
		}
		if vpcs[0].Status == VSwitchStatusAvailable {
			break
		}
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return getECSErrorFromString("Timeout")
		}
		time.Sleep(DefaultWaitForInterval * time.Second)
	}
	return nil
}
