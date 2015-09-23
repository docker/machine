package engine

type EngineOptions struct {
	ArbitraryFlags   []string
	Dns              []string
	GraphDir         string
	Env              []string
	Ipv6             bool
	InsecureRegistry []string
	Labels           []string
	LogLevel         string
	StorageDriver    string
	SelinuxEnabled   bool
	TlsVerify        bool
	RegistryMirror   []string
	InstallURL       string
}
