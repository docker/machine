package ecs

type DescribeInstanceTypesArgs struct {
}

type InstanceTypeItemType struct {
	InstanceTypeId string
	CpuCoreCount   int
	MemorySize     float64
}

type DescribeInstanceTypesResponse struct {
	CommonResponse
	InstanceTypes struct {
		InstanceType []InstanceTypeItemType
	}
}

// DescribeInstanceTypes describes all instance types
func (client *Client) DescribeInstanceTypes() (instanceTypes []InstanceTypeItemType, err error) {
	response := DescribeInstanceTypesResponse{}

	err = client.Invoke("DescribeInstanceTypes", &DescribeInstanceTypesArgs{}, &response)

	if err != nil {
		return []InstanceTypeItemType{}, err
	}
	return response.InstanceTypes.InstanceType, nil
}
