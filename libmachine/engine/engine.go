package engine

type EngineOptions struct {
	ArbitraryFlags   []string
	Dns              []string
	GraphDir         string
	Ipv6             bool
	InsecureRegistry []string
	Labels           []string
	LogLevel         string
	StorageDriver    string
	SelinuxEnabled   bool
	TlsCaCert        string
	TlsCert          string
	TlsKey           string
	TlsVerify        bool
	RegistryMirror   []string
}
