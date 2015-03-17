package amz

type IpPermission struct {
	IpProtocol string `xml:"ipProtocol"`
	FromPort   int    `xml:"fromPort"`
	ToPort     int    `xml:"toPort"`
	IpRange    string `xml:"ipRanges"`
}
