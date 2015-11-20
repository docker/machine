package auth

type Options struct {
	CertDir              string
	CaCertPath           string `yaml:"tls_ca_cert"`
	CaPrivateKeyPath     string `yaml:"tls_ca_key"`
	CaCertRemotePath     string
	ServerCertPath       string
	ServerKeyPath        string
	ClientKeyPath        string `yaml:"tls_client_key"`
	ServerCertRemotePath string
	ServerKeyRemotePath  string
	ClientCertPath       string `yaml:"tls_client_cert"`
	ServerCertSANs       []string
	// StorePath is left in for historical reasons, but not really meant to
	// be used directly.
	StorePath          string
	// SkipCertGeneration 
	SkipCertGeneration bool
}
