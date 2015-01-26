package amz

type DescribeKeyPairsResponse struct {
	RequestId string    `xml:"requestId"`
	KeySet    []KeyPair `xml:"keySet>item"`
}
