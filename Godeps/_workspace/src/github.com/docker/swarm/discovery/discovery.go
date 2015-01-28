package discovery

import (
	"errors"
	"fmt"
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type Node struct {
	Host string
	Port string
}

func NewNode(url string) (*Node, error) {
	host, port, err := net.SplitHostPort(url)
	if err != nil {
		return nil, err
	}
	return &Node{host, port}, nil
}

func (n Node) String() string {
	return fmt.Sprintf("%s:%s", n.Host, n.Port)
}

type WatchCallback func(nodes []*Node)

type DiscoveryService interface {
	Initialize(string, int) error
	Fetch() ([]*Node, error)
	Watch(WatchCallback)
	Register(string) error
}

var (
	discoveries       map[string]DiscoveryService
	ErrNotSupported   = errors.New("discovery service not supported")
	ErrNotImplemented = errors.New("not implemented in this discovery service")
)

func init() {
	discoveries = make(map[string]DiscoveryService)
}

func Register(scheme string, d DiscoveryService) error {
	if _, exists := discoveries[scheme]; exists {
		return fmt.Errorf("scheme already registered %s", scheme)
	}
	log.WithField("name", scheme).Debug("Registering discovery service")
	discoveries[scheme] = d

	return nil
}

func parse(rawurl string) (string, string) {
	parts := strings.SplitN(rawurl, "://", 2)

	// nodes:port,node2:port => nodes://node1:port,node2:port
	if len(parts) == 1 {
		return "nodes", parts[0]
	}
	return parts[0], parts[1]
}

func New(rawurl string, heartbeat int) (DiscoveryService, error) {
	scheme, uri := parse(rawurl)

	if discovery, exists := discoveries[scheme]; exists {
		log.WithFields(log.Fields{"name": scheme, "uri": uri}).Debug("Initializing discovery service")
		err := discovery.Initialize(uri, heartbeat)
		return discovery, err
	}

	return nil, ErrNotSupported
}
