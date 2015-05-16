package ecs

import (
	"github.com/denverdino/aliyungo/util"
)

type NicType string

const (
	NicTypeInternet = NicType("internet")
	NicTypeIntranet = NicType("intranet")
)

type IpProtocol string

const (
	IpProtocolAll  = IpProtocol("all")
	IpProtocolTCP  = IpProtocol("tcp")
	IpProtocolUDP  = IpProtocol("udp")
	IpProtocolICMP = IpProtocol("icmp")
	IpProtocolGRE  = IpProtocol("gre")
)

type PermissionPolicy string

const (
	PermissionPolicyAccept = PermissionPolicy("accept")
	PermissionPolicyDrop   = PermissionPolicy("drop")
)

type DescribeSecurityGroupAttributeArgs struct {
	SecurityGroupId string
	RegionId        Region
	NicType         NicType //enum for internet (default) |intranet
}

type PermissionType struct {
	IpProtocol              IpProtocol
	PortRange               string
	SourceCidrIp            string
	SourceGroupId           string
	SourceGroupOwnerAccount string
	Policy                  PermissionPolicy
	NicType                 NicType
}

type DescribeSecurityGroupAttributeResponse struct {
	CommonResponse

	SecurityGroupId   string
	SecurityGroupName string
	RegionId          Region
	Description       string
	Permissions       struct {
		Permission []PermissionType
	}
	VpcId string
}

func (client *Client) DescribeSecurityGroupAttribute(args *DescribeSecurityGroupAttributeArgs) (response *DescribeSecurityGroupAttributeResponse, err error) {
	response = &DescribeSecurityGroupAttributeResponse{}
	err = client.Invoke("DescribeSecurityGroupAttribute", args, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type DescribeSecurityGroupsArgs struct {
	RegionId Region
	VpcId    string
	Pagination
}

type SecurityGroupItemType struct {
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
	VpcId             string
	CreationTime      util.ISO6801Time
}

type DescribeSecurityGroupsResponse struct {
	CommonResponse

	PaginationResult
	RegionId       Region
	SecurityGroups struct {
		SecurityGroup []SecurityGroupItemType
	}
}

// DescribeSecurityGroups describes security groups
func (client *Client) DescribeSecurityGroups(args *DescribeSecurityGroupsArgs) (securityGroupItems []SecurityGroupItemType, pagination *PaginationResult, err error) {
	args.validate()
	response := DescribeSecurityGroupsResponse{}

	err = client.Invoke("DescribeSecurityGroups", args, &response)

	if err != nil {
		return nil, nil, err
	}

	return response.SecurityGroups.SecurityGroup, &response.PaginationResult, nil
}

type CreateSecurityGroupArgs struct {
	RegionId          Region
	SecurityGroupName string
	Description       string
	VpcId             string
	ClientToken       string
}

type CreateSecurityGroupResponse struct {
	CommonResponse

	SecurityGroupId string
}

// CreateSecurityGroup creates security group
func (client *Client) CreateSecurityGroup(args *CreateSecurityGroupArgs) (securityGroupId string, err error) {
	response := CreateSecurityGroupResponse{}
	err = client.Invoke("CreateSecurityGroup", args, &response)
	if err != nil {
		return "", err
	}
	return response.SecurityGroupId, err
}

type DeleteSecurityGroupArgs struct {
	RegionId        Region
	SecurityGroupId string
}

type DeleteSecurityGroupResponse struct {
	CommonResponse
}

// DeleteSecurityGroup deletes security group
func (client *Client) DeleteSecurityGroup(regionId Region, securityGroupId string) error {
	args := DeleteSecurityGroupArgs{
		RegionId:        regionId,
		SecurityGroupId: securityGroupId,
	}
	response := DeleteSecurityGroupResponse{}
	err := client.Invoke("DeleteSecurityGroup", &args, &response)
	return err
}

type ModifySecurityGroupAttributeArgs struct {
	RegionId          Region
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
}

type ModifySecurityGroupAttributeResponse struct {
	CommonResponse
}

// ModifySecurityGroupAttribute modifies attribute of security group
func (client *Client) ModifySecurityGroupAttribute(args *ModifySecurityGroupAttributeArgs) error {
	response := ModifySecurityGroupAttributeResponse{}
	err := client.Invoke("ModifySecurityGroupAttribute", args, &response)
	return err
}

type AuthorizeSecurityGroupArgs struct {
	SecurityGroupId         string
	RegionId                Region
	IpProtocol              IpProtocol
	PortRange               string
	SourceGroupId           string
	SourceGroupOwnerAccount string
	SourceCidrIp            string           // IPv4 only, default 0.0.0.0/0
	Policy                  PermissionPolicy // enum of accept (default) | drop
	Priority                int              // 1 - 100, default 1
	NicType                 NicType          // enum of internet | intranet (default)
}

type AuthorizeSecurityGroupResponse struct {
	CommonResponse
}

// AuthorizeSecurityGroup authorize permissions to security group
func (client *Client) AuthorizeSecurityGroup(args *AuthorizeSecurityGroupArgs) error {
	response := AuthorizeSecurityGroupResponse{}
	err := client.Invoke("AuthorizeSecurityGroup", args, &response)
	return err
}
