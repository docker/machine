package engine

type EngineOptions struct {
	Dns            []string
	GraphDir       string
	Ipv6           bool
	Labels         []string
	LogLevel       string
	StorageDriver  string
	SelinuxEnabled bool
	TlsCaCert      string
	TlsCert        string
	TlsKey         string
	TlsVerify      bool
	RegistryMirror []string
}
