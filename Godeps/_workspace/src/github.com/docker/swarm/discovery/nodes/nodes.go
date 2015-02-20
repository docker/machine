package nodes

import (
	"strings"

	"github.com/docker/swarm/discovery"
)

type NodesDiscoveryService struct {
	nodes []*discovery.Node
}

func init() {
	discovery.Register("nodes", &NodesDiscoveryService{})
}

func (s *NodesDiscoveryService) Initialize(uris string, _ int) error {
	for _, ip := range strings.Split(uris, ",") {
		node, err := discovery.NewNode(ip)
		if err != nil {
			return err
		}
		s.nodes = append(s.nodes, node)
	}

	return nil
}
func (s *NodesDiscoveryService) Fetch() ([]*discovery.Node, error) {
	return s.nodes, nil
}

func (s *NodesDiscoveryService) Watch(callback discovery.WatchCallback) {
}

func (s *NodesDiscoveryService) Register(addr string) error {
	return discovery.ErrNotImplemented
}
