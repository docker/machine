// API on Network

package ecs

import (
	"github.com/denverdino/aliyungo/util"
	"time"
)

type AllocatePublicIpAddressArgs struct {
	InstanceId string
}

type AllocatePublicIpAddressResponse struct {
	CommonResponse

	IpAddress string
}

// AllocatePublicIpAddress allocates Public Ip Address
func (client *Client) AllocatePublicIpAddress(instanceId string) (ipAddress string, err error) {
	args := AllocatePublicIpAddressArgs{
		InstanceId: instanceId,
	}
	response := AllocatePublicIpAddressResponse{}
	err = client.Invoke("AllocatePublicIpAddress", &args, &response)
	if err != nil {
		return "", err
	}
	return response.IpAddress, nil
}

type ModifyInstanceNetworkSpec struct {
	InstanceId              string
	InternetMaxBandwidthOut *int
	InternetMaxBandwidthIn  *int
}

type ModifyInstanceNetworkSpecResponse struct {
	CommonResponse
}

// ModifyInstanceNetworkSpec modifies instance network spec
func (client *Client) ModifyInstanceNetworkSpec(args *ModifyInstanceNetworkSpec) error {

	response := ModifyInstanceNetworkSpecResponse{}
	return client.Invoke("ModifyInstanceNetworkSpec", args, &response)
}

type AllocateEipAddressArgs struct {
	RegionId           Region
	Bandwidth          int
	InternetChargeType InternetChargeType
	ClientToken        string
}

type AllocateEipAddressResponse struct {
	CommonResponse
	EipAddress   string
	AllocationId string
}

// AllocateEipAddress allocates Eip Address
func (client *Client) AllocateEipAddress(args *AllocateEipAddressArgs) (EipAddress string, AllocationId string, err error) {
	if args.Bandwidth == 0 {
		args.Bandwidth = 5
	}
	response := AllocateEipAddressResponse{}
	err = client.Invoke("AllocateEipAddress", args, &response)
	if err != nil {
		return "", "", err
	}
	return response.EipAddress, response.AllocationId, nil
}

type AssociateEipAddressArgs struct {
	AllocationId string
	InstanceId   string
}

type AssociateEipAddressResponse struct {
	CommonResponse
}

// AssociateEipAddress associates EIP address to VM instance
func (client *Client) AssociateEipAddress(allocationId string, instanceId string) error {
	args := AssociateEipAddressArgs{
		AllocationId: allocationId,
		InstanceId:   instanceId,
	}
	response := ModifyInstanceNetworkSpecResponse{}
	return client.Invoke("AssociateEipAddress", &args, &response)
}

// Status of disks
type EipStatus string

const (
	EipStatusAssociating   = EipStatus("Associating")
	EipStatusUnassociating = EipStatus("Unassociating")
	EipStatusInUse         = EipStatus("InUse")
	EipStatusAvailable     = EipStatus("Available")
)

type DescribeEipAddressesArgs struct {
	RegionId     Region
	Status       EipStatus //enum Associating | Unassociating | InUse | Available
	EipAddress   string
	AllocationId string
	Pagination
}

type EipAddressSetType struct {
	RegionId           Region
	IpAddress          string
	AllocationId       string
	Status             EipStatus
	InstanceId         string
	Bandwidth          string // Why string
	InternetChargeType InternetChargeType
	OperationLocks     OperationLocksType
	AllocationTime     util.ISO6801Time
}

type DescribeEipAddressesResponse struct {
	CommonResponse
	PaginationResult
	EipAddresses struct {
		EipAddress []EipAddressSetType
	}
}

// DescribeInstanceStatus describes instance status
func (client *Client) DescribeEipAddresses(args *DescribeEipAddressesArgs) (eipAddresses []EipAddressSetType, pagination *PaginationResult, err error) {
	args.validate()
	response := DescribeEipAddressesResponse{}

	err = client.Invoke("DescribeEipAddresses", args, &response)

	if err == nil {
		return response.EipAddresses.EipAddress, &response.PaginationResult, nil
	}

	return nil, nil, err
}

type ModifyEipAddressAttributeArgs struct {
	AllocationId string
	Bandwidth    int
}

type ModifyEipAddressAttributeResponse struct {
	CommonResponse
}

// ModifyEipAddressAttribute Modifies EIP attribute
func (client *Client) ModifyEipAddressAttribute(allocationId string, bandwidth int) error {
	args := ModifyEipAddressAttributeArgs{
		AllocationId: allocationId,
		Bandwidth:    bandwidth,
	}
	response := ModifyEipAddressAttributeResponse{}
	return client.Invoke("ModifyEipAddressAttribute", &args, &response)
}

type UnallocateEipAddressArgs struct {
	AllocationId string
	InstanceId   string
}

type UnallocateEipAddressResponse struct {
	CommonResponse
}

// UnassociateEipAddress unallocates Eip Address from instance
func (client *Client) UnassociateEipAddress(allocationId string, instanceId string) error {
	args := UnallocateEipAddressArgs{
		AllocationId: allocationId,
		InstanceId:   instanceId,
	}
	response := UnallocateEipAddressResponse{}
	return client.Invoke("UnassociateEipAddress", &args, &response)
}

type ReleaseEipAddressArgs struct {
	AllocationId string
}

type ReleaseEipAddressResponse struct {
	CommonResponse
}

// ReleaseEipAddress releases Eip address
func (client *Client) ReleaseEipAddress(allocationId string) error {
	args := ReleaseEipAddressArgs{
		AllocationId: allocationId,
	}
	response := ReleaseEipAddressResponse{}
	return client.Invoke("ReleaseEipAddress", &args, &response)
}

// WaitForVSwitchAvailable waits for VSwitch to given status
func (client *Client) WaitForEip(regionId Region, allocationId string, status EipStatus, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	args := DescribeEipAddressesArgs{
		RegionId:     regionId,
		AllocationId: allocationId,
	}
	for {
		vpcs, _, err := client.DescribeEipAddresses(&args)
		if err != nil {
			return err
		}
		if vpcs[0].Status == status {
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
