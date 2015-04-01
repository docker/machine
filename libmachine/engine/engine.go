package engine

type EngineOptions struct {
	DNS            []string
	GraphDir       string
	Ipv6           bool
	Labels         []string
	LogLevel       string
	StorageDriver  string
	SelinuxEnabled bool
	TLSCaCert      string
	TLSCert        string
	TLSKey         string
	TLSVerify      bool
	RegistryMirror []string
}
