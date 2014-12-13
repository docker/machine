package amz

type DescribeInstancesResponse struct {
	RequestId      string `xml:"requestId"`
	ReservationSet []struct {
		InstancesSet []EC2Instance `xml:"instancesSet>item"`
	} `xml:"reservationSet>item"`
}
