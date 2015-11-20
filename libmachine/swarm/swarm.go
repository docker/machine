package swarm

const (
	DiscoveryServiceEndpoint = "https://discovery-stage.hub.docker.com/v1"
)

type Options struct {
	IsSwarm        bool
	Address        string `yaml:"addr"`
	Discovery      string `yaml:"discovery"`
	Master         bool   `yaml:"master"`
	Host           string `yaml:"host"`
	Image          string `yaml:"image"`
	Strategy       string `yaml:"strategy"`
	Heartbeat      int
	Overcommit     float64
	ArbitraryFlags []string `yaml:"opt"`
	Env            []string
}
