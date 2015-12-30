package engine

type Options struct {
	ArbitraryFlags   []string `yaml:"opt"`
	DNS              []string `json:"Dns"`
	GraphDir         string   `yaml:"graph_dir"`
	Env              []string `yaml:"env"`
	Ipv6             bool     `yaml:"ipv6"`
	InsecureRegistry []string `yaml:"insecure_registry"`
	Labels           []string `yaml:"label"`
	LogLevel         string   `yaml:"storage_driver"`
	StorageDriver    string   `yaml:"storage_driver"`
	SelinuxEnabled   bool     `yaml:"selinux_enabled"`
	TLSVerify        bool     `json:"TlsVerify"`
	RegistryMirror   []string `yaml:"registry_mirror"`
	InstallURL       string   `yaml:"install_url"`
}
