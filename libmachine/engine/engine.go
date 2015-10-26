package engine

type Options struct {
	ArbitraryFlags   []string
	DNS              []string
	GraphDir         string
	Env              []string
	Ipv6             bool
	InsecureRegistry []string
	Labels           []string
	LogLevel         string
	StorageDriver    string
	SelinuxEnabled   bool
	TLSVerify        bool
	RegistryMirror   []string
	InstallURL       string
}
