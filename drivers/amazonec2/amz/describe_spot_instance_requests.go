package amz

type DescribeSpotInstanceRequestsResponse struct {
	RequestId              string `xml:"requestId"`
	SpotInstanceRequestSet []struct {
		Status struct {
			Code string `xml:"code"`
		} `xml:"status"`
		InstanceId string `xml:"instanceId"`
	} `xml:"spotInstanceRequestSet>item"`
}
