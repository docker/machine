package amz

type CreateKeyPairResponse struct {
	KeyName        string `xml:"keyName"`
	KeyFingerprint string `xml:"keyFingerprint"`
	KeyMaterial    []byte `xml:"keyMaterial"`
}

type ImportKeyPairResponse struct {
	KeyName        string `xml:"keyName"`
	KeyFingerprint string `xml:"keyFingerprint"`
	KeyMaterial    []byte `xml:"keyMaterial"`
}

type KeyPair struct {
	KeyFingerprint string `xml:"keyFingerprint"`
	KeyName        string `xml:"keyName"`
}
