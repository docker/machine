package swarm

const (
	DockerImage              = "swarm:latest"
	DiscoveryServiceEndpoint = "https://discovery-stage.hub.docker.com/v1"
)

type SwarmOptions struct {
	IsSwarm    bool
	Address    string
	Discovery  string
	Master     bool
	Host       string
	Strategy   string
	Heartbeat  int
	Overcommit float64
	TlsCaCert  string
	TlsCert    string
	TlsKey     string
	TlsVerify  bool
}
