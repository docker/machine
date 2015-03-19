package swarm

type SwarmOptions struct {
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
