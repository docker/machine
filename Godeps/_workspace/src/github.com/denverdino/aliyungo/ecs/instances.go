package ecs

import (
	"time"

	"github.com/denverdino/aliyungo/util"
)

// InstanceStatus represents instance status
type InstanceStatus string

// Constants of InstanceStatus
const (
	Creating = InstanceStatus("Creating")
	Running  = InstanceStatus("Running")
	Starting = InstanceStatus("Starting")

	Stopped  = InstanceStatus("Stopped")
	Stopping = InstanceStatus("Stopping")
)

type InternetChargeType string

const (
	PayByBandwidth = InternetChargeType("PayByBandwidth")
	PayByTraffic   = InternetChargeType("PayByTraffic")
)

type LockReason string

const (
	LockReasonFinancial = LockReason("financial")
	LockReasonSecurity  = LockReason("security")
)

type DescribeInstanceStatusArgs struct {
	RegionId Region
	ZoneId   string
	Pagination
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&instancestatusitemtype
type InstanceStatusItemType struct {
	InstanceId string
	Status     InstanceStatus
}

type DescribeInstanceStatusResponse struct {
	CommonResponse
	PaginationResult
	InstanceStatuses struct {
		InstanceStatus []InstanceStatusItemType
	}
}

// DescribeInstanceStatus describes instance status
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&describeinstancestatus
func (client *Client) DescribeInstanceStatus(args *DescribeInstanceStatusArgs) (instanceStatuses []InstanceStatusItemType, pagination *PaginationResult, err error) {
	args.validate()
	response := DescribeInstanceStatusResponse{}

	err = client.Invoke("DescribeInstanceStatus", args, &response)

	if err == nil {
		return response.InstanceStatuses.InstanceStatus, &response.PaginationResult, nil
	}

	return nil, nil, err
}

type StopInstanceArgs struct {
	InstanceId string
	ForceStop  bool
}

type StopInstanceResponse struct {
	CommonResponse
}

// StopInstance stops instance
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&stopinstance
func (client *Client) StopInstance(instanceId string, forceStop bool) error {
	args := StopInstanceArgs{
		InstanceId: instanceId,
		ForceStop:  forceStop,
	}
	response := StopInstanceResponse{}
	err := client.Invoke("StopInstance", &args, &response)
	return err
}

type StartInstanceArgs struct {
	InstanceId string
}

type StartInstanceResponse struct {
	CommonResponse
}

// StartInstance starts instance
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&startinstance
func (client *Client) StartInstance(instanceId string) error {
	args := StartInstanceArgs{InstanceId: instanceId}
	response := StartInstanceResponse{}
	err := client.Invoke("StartInstance", &args, &response)
	return err
}

type RebootInstanceArgs struct {
	InstanceId string
	ForceStop  bool
}

type RebootInstanceResponse struct {
	CommonResponse
}

// RebootInstance reboot instance
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&rebootinstance
func (client *Client) RebootInstance(instanceId string, forceStop bool) error {
	request := RebootInstanceArgs{
		InstanceId: instanceId,
		ForceStop:  forceStop,
	}
	response := RebootInstanceResponse{}
	err := client.Invoke("RebootInstance", &request, &response)
	return err
}

type DescribeInstanceAttributeArgs struct {
	InstanceId string
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&operationlockstype
type OperationLocksType struct {
	LockReason []LockReason //enum for financial, security
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&securitygroupidsettype
type SecurityGroupIdSetType struct {
	SecurityGroupId string
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&ipaddresssettype
type IpAddressSetType struct {
	IpAddress []string
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&vpcattributestype
type VpcAttributesType struct {
	VpcId            string
	VSwitchId        string
	PrivateIpAddress IpAddressSetType
	NatIpAddress     string
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&eipaddressassociatetype
type EipAddressAssociateType struct {
	AllocationId       string
	IpAddress          string
	Bandwidth          int
	InternetChargeType InternetChargeType
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&instanceattributestype
type InstanceAttributesType struct {
	InstanceId       string
	InstanceName     string
	Description      string
	ImageId          string
	RegionId         Region
	ZoneId           string
	ClusterId        string
	InstanceType     string
	HostName         string
	Status           InstanceStatus
	OperationLocks   OperationLocksType
	SecurityGroupIds struct {
		SecurityGroupId []string
	}
	PublicIpAddress         IpAddressSetType
	InnerIpAddress          IpAddressSetType
	InstanceNetworkType     string //enum Classic | Vpc
	InternetMaxBandwidthIn  int
	InternetMaxBandwidthOut int
	InternetChargeType      InternetChargeType
	CreationTime            util.ISO6801Time //time.Time
	VpcAttributes           VpcAttributesType
	EipAddress              EipAddressAssociateType
}

type DescribeInstanceAttributeResponse struct {
	CommonResponse
	InstanceAttributesType
}

// DescribeInstanceAttribute describes instance attribute
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&describeinstanceattribute
func (client *Client) DescribeInstanceAttribute(instanceId string) (instance *InstanceAttributesType, err error) {
	args := DescribeInstanceAttributeArgs{InstanceId: instanceId}

	response := DescribeInstanceAttributeResponse{}
	err = client.Invoke("DescribeInstanceAttribute", &args, &response)
	if err != nil {
		return nil, err
	}
	return &response.InstanceAttributesType, err
}

// Default timeout value for WaitForInstance method
const InstanceDefaultTimeout = 120

// WaitForInstance waits for instance to given status
func (client *Client) WaitForInstance(instanceId string, status InstanceStatus, timeout int) error {
	if timeout <= 0 {
		timeout = InstanceDefaultTimeout
	}
	for {
		instance, err := client.DescribeInstanceAttribute(instanceId)
		if err != nil {
			return err
		}
		if instance.Status == status {
			//TODO
			//Sleep one more time for timing issues
			time.Sleep(DefaultWaitForInterval * time.Second)
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

type DescribeInstanceVncUrlArgs struct {
	RegionId   Region
	InstanceId string
}

type DescribeInstanceVncUrlResponse struct {
	CommonResponse
	VncUrl string
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&describeinstancevncurl
func (client *Client) DescribeInstanceVncUrl(args *DescribeInstanceVncUrlArgs) (string, error) {
	response := DescribeInstanceVncUrlResponse{}

	err := client.Invoke("DescribeInstanceVncUrl", args, &response)

	if err == nil {
		return response.VncUrl, nil
	}

	return "", err
}

type DescribeInstancesArgs struct {
	RegionId            Region
	VpcId               string
	VSwitchId           string
	ZoneId              string
	InstanceIds         string
	InstanceNetworkType string
	PrivateIpAddresses  string
	InnerIpAddresses    string
	PublicIpAddresses   string
	SecurityGroupId     string
	Pagination
}

type DescribeInstancesResponse struct {
	CommonResponse
	PaginationResult
	Instances struct {
		Instance []InstanceAttributesType
	}
}

// DescribeInstances describes instances
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&describeinstances
func (client *Client) DescribeInstances(args *DescribeInstancesArgs) (instances []InstanceAttributesType, pagination *PaginationResult, err error) {
	args.validate()
	response := DescribeInstancesResponse{}

	err = client.Invoke("DescribeInstances", args, &response)

	if err == nil {
		return response.Instances.Instance, &response.PaginationResult, nil
	}

	return nil, nil, err
}

type DeleteInstanceArgs struct {
	InstanceId string
}

type DeleteInstanceResponse struct {
	CommonResponse
}

// DeleteInstance deletes instance
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&deleteinstance
func (client *Client) DeleteInstance(instanceId string) error {
	args := DeleteInstanceArgs{InstanceId: instanceId}
	response := DeleteInstanceResponse{}
	err := client.Invoke("DeleteInstance", &args, &response)
	return err
}

type DataDiskType struct {
	Size               int
	Category           DiskCategory //Enum cloud, ephemeral, ephemeral_ssd
	SnapshotId         string
	DiskName           string
	Description        string
	Device             string
	DeleteWithInstance bool
}

type SystemDiskType struct {
	Category    DiskCategory //Enum cloud, ephemeral, ephemeral_ssd
	DiskName    string
	Description string
}

type CreateInstanceArgs struct {
	RegionId                Region
	ZoneId                  string
	ImageId                 string
	InstanceType            string
	SecurityGroupId         string
	InstanceName            string
	Description             string
	InternetChargeType      InternetChargeType
	InternetMaxBandwidthIn  int
	InternetMaxBandwidthOut int
	HostName                string
	Password                string
	SystemDisk              SystemDiskType
	DataDisk                []DataDiskType
	VSwitchId               string
	PrivateIpAddress        string
	ClientToken             string
}

type CreateInstanceResponse struct {
	CommonResponse
	InstanceId string
}

// CreateInstance creates instance
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/instance&createinstance
func (client *Client) CreateInstance(args *CreateInstanceArgs) (instanceId string, err error) {
	response := CreateInstanceResponse{}
	err = client.Invoke("CreateInstance", args, &response)
	if err != nil {
		return "", err
	}
	return response.InstanceId, err
}
