package amz

type CreateSecurityGroupResponse struct {
	RequestId string `xml:"requestId"`
	Return    bool   `xml:"return"`
	GroupId   string `xml:"groupId"`
}

type DeleteSecurityGroupResponse struct {
	RequestId string `xml:"requestId"`
	Return    bool   `xml:"return"`
}

type SecurityGroup struct {
	GroupId string
	VpcId   string
}
