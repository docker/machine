package file

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/swarm/discovery"
)

type FileDiscoveryService struct {
	heartbeat int
	path      string
}

func init() {
	discovery.Register("file", &FileDiscoveryService{})
}

func (s *FileDiscoveryService) Initialize(path string, heartbeat int) error {
	s.path = path
	s.heartbeat = heartbeat
	return nil
}

func (s *FileDiscoveryService) Fetch() ([]*discovery.Node, error) {
	data, err := ioutil.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	var nodes []*discovery.Node

	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			node, err := discovery.NewNode(line)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (s *FileDiscoveryService) Watch(callback discovery.WatchCallback) {
	for _ = range time.Tick(time.Duration(s.heartbeat) * time.Second) {
		nodes, err := s.Fetch()
		if err == nil {
			callback(nodes)
		}
	}
}

func (s *FileDiscoveryService) Register(addr string) error {
	return discovery.ErrNotImplemented
}
