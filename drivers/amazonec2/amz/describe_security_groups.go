package amz

type DescribeSecurityGroupsResponse struct {
	RequestId         string          `xml:"requestId"`
	SecurityGroupInfo []SecurityGroup `xml:"securityGroupInfo>item"`
}
