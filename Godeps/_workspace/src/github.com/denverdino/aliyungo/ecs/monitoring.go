package ecs

import (
	"github.com/denverdino/aliyungo/util"
)

type DescribeInstanceMonitorDataArgs struct {
	InstanceId string
	StartTime  util.ISO6801Time
	EndTime    util.ISO6801Time
	Period     int //Default 60s
}

type InstanceMonitorDataType struct {
	InstanceId        string
	CPU               int
	IntranetRX        int
	IntranetTX        int
	IntranetBandwidth int
	InternetRX        int
	InternetTX        int
	InternetBandwidth int
	IOPSRead          int
	IOPSWrite         int
	BPSRead           int
	BPSWrite          int
	TimeStamp         util.ISO6801Time
}

type DescribeInstanceMonitorDataResponse struct {
	CommonResponse
	MonitorData struct {
		InstanceMonitorData []InstanceMonitorDataType
	}
}

// DescribeInstanceMonitorData describes instance monitoring data
func (client *Client) DescribeInstanceMonitorData(args *DescribeInstanceMonitorDataArgs) (monitorData []InstanceMonitorDataType, err error) {
	if args.Period == 0 {
		args.Period = 60
	}
	response := DescribeInstanceMonitorDataResponse{}
	err = client.Invoke("DescribeInstanceMonitorData", args, &response)
	if err != nil {
		return nil, err
	}
	return response.MonitorData.InstanceMonitorData, err
}

type DescribeEipMonitorDataArgs struct {
	AllocationId string
	StartTime    util.ISO6801Time
	EndTime      util.ISO6801Time
	Period       int //Default 60s
}

type EipMonitorDataType struct {
	EipRX        int
	EipTX        int
	EipFlow      int
	EipBandwidth int
	EipPackets   int
	TimeStamp    util.ISO6801Time
}

type DescribeEipMonitorDataResponse struct {
	CommonResponse
	EipMonitorDatas struct {
		EipMonitorData []EipMonitorDataType
	}
}

// DescribeEipMonitorData describes EIP monitoring data
func (client *Client) DescribeEipMonitorData(args *DescribeEipMonitorDataArgs) (monitorData []EipMonitorDataType, err error) {
	if args.Period == 0 {
		args.Period = 60
	}
	response := DescribeEipMonitorDataResponse{}
	err = client.Invoke("DescribeEipMonitorData", args, &response)
	if err != nil {
		return nil, err
	}
	return response.EipMonitorDatas.EipMonitorData, err
}

type DescribeDiskMonitorDataArgs struct {
	DiskId    string
	StartTime util.ISO6801Time
	EndTime   util.ISO6801Time
	Period    int //Default 60s
}

type DiskMonitorDataType struct {
	DiskId    string
	IOPSRead  int
	IOPSWrite int
	IOPSTotal int
	BPSRead   int
	BPSWrite  int
	BPSTotal  int
	TimeStamp util.ISO6801Time
}

type DescribeDiskMonitorDataResponse struct {
	CommonResponse
	TotalCount  int
	MonitorData struct {
		DiskMonitorData []DiskMonitorDataType
	}
}

// DescribeDiskMonitorData describes disk monitoring data
func (client *Client) DescribeDiskMonitorData(args *DescribeDiskMonitorDataArgs) (monitorData []DiskMonitorDataType, totalCount int, err error) {
	if args.Period == 0 {
		args.Period = 60
	}
	response := DescribeDiskMonitorDataResponse{}
	err = client.Invoke("DescribeDiskMonitorData", args, &response)
	if err != nil {
		return nil, 0, err
	}
	return response.MonitorData.DiskMonitorData, response.TotalCount, err
}
