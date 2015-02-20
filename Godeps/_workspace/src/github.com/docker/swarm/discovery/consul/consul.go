package consul

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	consul "github.com/armon/consul-api"
	"github.com/docker/swarm/discovery"
)

type ConsulDiscoveryService struct {
	heartbeat time.Duration
	client    *consul.Client
	prefix    string
	lastIndex uint64
}

func init() {
	discovery.Register("consul", &ConsulDiscoveryService{})
}

func (s *ConsulDiscoveryService) Initialize(uris string, heartbeat int) error {
	parts := strings.SplitN(uris, "/", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid format %q, missing <path>", uris)
	}
	addr := parts[0]
	path := parts[1]

	config := consul.DefaultConfig()
	config.Address = addr

	client, err := consul.NewClient(config)
	if err != nil {
		return err
	}
	s.client = client
	s.heartbeat = time.Duration(heartbeat) * time.Second
	s.prefix = path + "/"
	kv := s.client.KV()
	p := &consul.KVPair{Key: s.prefix, Value: nil}
	if _, err = kv.Put(p, nil); err != nil {
		return err
	}
	_, meta, err := kv.Get(s.prefix, nil)
	if err != nil {
		return err
	}
	s.lastIndex = meta.LastIndex
	return nil
}
func (s *ConsulDiscoveryService) Fetch() ([]*discovery.Node, error) {
	kv := s.client.KV()
	pairs, _, err := kv.List(s.prefix, nil)
	if err != nil {
		return nil, err
	}

	var nodes []*discovery.Node

	for _, pair := range pairs {
		if pair.Key == s.prefix {
			continue
		}
		node, err := discovery.NewNode(string(pair.Value))
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (s *ConsulDiscoveryService) Watch(callback discovery.WatchCallback) {
	for _ = range s.waitForChange() {
		log.WithField("name", "consul").Debug("Discovery watch triggered")
		nodes, err := s.Fetch()
		if err == nil {
			callback(nodes)
		}
	}
}

func (s *ConsulDiscoveryService) Register(addr string) error {
	kv := s.client.KV()
	p := &consul.KVPair{Key: path.Join(s.prefix, addr), Value: []byte(addr)}
	_, err := kv.Put(p, nil)
	return err
}

func (s *ConsulDiscoveryService) waitForChange() <-chan uint64 {
	c := make(chan uint64)
	go func() {
		for {
			kv := s.client.KV()
			option := &consul.QueryOptions{
				WaitIndex: s.lastIndex,
				WaitTime:  s.heartbeat}
			_, meta, err := kv.List(s.prefix, option)
			if err != nil {
				log.WithField("name", "consul").Errorf("Discovery error: %v", err)
				break
			}
			s.lastIndex = meta.LastIndex
			c <- s.lastIndex
		}
		close(c)
	}()
	return c
}
