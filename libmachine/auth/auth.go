package auth

type AuthOptions struct {
	CertDir              string
	CaCertPath           string
	CaPrivateKeyPath     string
	CaCertRemotePath     string
	ServerCertPath       string
	ServerKeyPath        string
	ClientKeyPath        string
	ServerCertRemotePath string
	ServerKeyRemotePath  string
	ClientCertPath       string

	// StorePath is left in for historical reasons, but not really meant to
	// be used directly.
	StorePath string
}
