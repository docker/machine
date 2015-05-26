package ecs

import (
	"github.com/denverdino/aliyungo/util"
)

type DescribeVRoutersArgs struct {
	VRouterId string
	RegionId  Region
	Pagination
}

type VRouterSetType struct {
	VRouterId     string
	RegionId      Region
	VpcId         string
	RouteTableIds struct {
		RouteTableId []string
	}
	VRouterName  string
	Description  string
	CreationTime util.ISO6801Time
}

type DescribeVRoutersResponse struct {
	CommonResponse
	PaginationResult
	VRouters struct {
		VRouter []VRouterSetType
	}
}

// DescribeVRouters describes Virtual Routers
func (client *Client) DescribeVRouters(args *DescribeVRoutersArgs) (vrouters []VRouterSetType, pagination *PaginationResult, err error) {
	args.validate()
	response := DescribeVRoutersResponse{}

	err = client.Invoke("DescribeVRouters", args, &response)

	if err == nil {
		return response.VRouters.VRouter, &response.PaginationResult, nil
	}

	return nil, nil, err
}

type ModifyVRouterAttributeArgs struct {
	VRouterId   string
	VRouterName string
	Description string
}

type ModifyVRouterAttributeResponse struct {
	CommonResponse
}

// ModifyVRouterAttribute modifies attribute of Virtual Router
func (client *Client) ModifyVRouterAttribute(args *ModifyVRouterAttributeArgs) error {
	response := ModifyVRouterAttributeResponse{}
	return client.Invoke("ModifyVRouterAttribute", args, &response)
}
