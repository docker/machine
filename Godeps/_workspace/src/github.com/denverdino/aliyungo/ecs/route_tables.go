package ecs

import (
	"github.com/denverdino/aliyungo/util"
	"time"
)

type DescribeRouteTablesArgs struct {
	VRouterId    string
	RouteTableId string
	Pagination
}

type RouteTableType string

const (
	RouteTableSystem = RouteTableType("System")
	RouteTableCustom = RouteTableType("Custom")
)

type RouteEntryStatus string

const (
	RouteEntryStatusPending   = RouteEntryStatus("Pending")
	RouteEntryStatusAvailable = RouteEntryStatus("Available")
	RouteEntryStatusModifying = RouteEntryStatus("Modifying")
)

type RouteEntrySetType struct {
	RouteTableId         string
	DestinationCidrBlock string
	Type                 RouteTableType
	NextHopId            string
	Status               RouteEntryStatus // enum Pending | Available | Modifying
}

type RouteTableSetType struct {
	VRouterId    string
	RouteTableId string
	RouteEntrys  struct {
		RouteEntry []RouteEntrySetType
	}
	RouteTableType RouteTableType
	CreationTime   util.ISO6801Time
}

type DescribeRouteTablesResponse struct {
	CommonResponse
	PaginationResult
	RouteTables struct {
		RouteTable []RouteTableSetType
	}
}

// DescribeRouteTables describes Virtual Routers
func (client *Client) DescribeRouteTables(args *DescribeRouteTablesArgs) (routeTables []RouteTableSetType, pagination *PaginationResult, err error) {
	args.validate()
	response := DescribeRouteTablesResponse{}

	err = client.Invoke("DescribeRouteTables", args, &response)

	if err == nil {
		return response.RouteTables.RouteTable, &response.PaginationResult, nil
	}

	return nil, nil, err
}

type NextHopType string

const (
	NextHopIntance = NextHopType("Instance") //Default
	NextHopTunnel  = NextHopType("Tunnel")
)

type CreateRouteEntryArgs struct {
	RouteTableId         string
	DestinationCidrBlock string
	NextHopType          NextHopType
	NextHopId            string
	ClientToken          string
}

type CreateRouteEntryResponse struct {
	CommonResponse
}

// CreateRouteEntry creates route entry
func (client *Client) CreateRouteEntry(args *CreateRouteEntryArgs) error {
	response := CreateRouteEntryResponse{}
	return client.Invoke("CreateRouteEntry", args, &response)
}

type DeleteRouteEntryArgs struct {
	RouteTableId         string
	DestinationCidrBlock string
	NextHopId            string
}

type DeleteRouteEntryResponse struct {
	CommonResponse
}

// DeleteRouteEntry deletes route entry
func (client *Client) DeleteRouteEntry(args *DeleteRouteEntryArgs) error {
	response := DeleteRouteEntryResponse{}
	return client.Invoke("DeleteRouteEntry", args, &response)
}

// WaitForAllRouteEntriesAvailable waits for all route entries to Available status
func (client *Client) WaitForAllRouteEntriesAvailable(vrouterId string, routeTableId string, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	args := DescribeRouteTablesArgs{
		VRouterId:    vrouterId,
		RouteTableId: routeTableId,
	}
	for {

		routeTables, _, err := client.DescribeRouteTables(&args)

		if err != nil {
			return err
		}
		sucess := true

	loop:
		for _, routeTable := range routeTables {
			for _, routeEntry := range routeTable.RouteEntrys.RouteEntry {
				if routeEntry.Status != RouteEntryStatusAvailable {
					sucess = false
					break loop
				}
			}
		}
		if sucess {
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
