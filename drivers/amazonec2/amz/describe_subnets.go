package amz

type DescribeSubnetsResponse struct {
	RequestId string   `xml:"requestId"`
	SubnetSet []Subnet `xml:"subnetSet>item"`
}

type Subnet struct {
	SubnetId         string `xml:"subnetId"`
	State            string `xml:"state"`
	VpcId            string `xml:"vpcId"`
	CidrBlock        string `xml:"cidrBlock"`
	AvailabilityZone string `xml:"availabilityZone"`
	DefaultForAz     bool   `xml:"defaultForAz"`
}
