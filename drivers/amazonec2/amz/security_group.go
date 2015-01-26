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
	GroupName           string         `xml:"groupName"`
	GroupId             string         `xml:"groupId"`
	VpcId               string         `xml:"vpcId"`
	OwnerId             string         `xml:"ownerId"`
	IpPermissions       []IpPermission `xml:"ipPermissions>item,omitempty"`
	IpPermissionsEgress []IpPermission `xml:"ipPermissionsEgress>item,omitempty"`
}
