package cert

type CertPathInfo struct {
	CaCertPath       string
	CaPrivateKeyPath string
	ClientCertPath   string
	ClientKeyPath    string
	ServerCertPath   string
	ServerKeyPath    string
}
