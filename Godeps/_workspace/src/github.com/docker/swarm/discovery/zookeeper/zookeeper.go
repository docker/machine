package zookeeper

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/discovery"
	"github.com/samuel/go-zookeeper/zk"
)

type ZkDiscoveryService struct {
	conn      *zk.Conn
	path      []string
	heartbeat int
}

func init() {
	discovery.Register("zk", &ZkDiscoveryService{})
}

func (s *ZkDiscoveryService) fullpath() string {
	return "/" + strings.Join(s.path, "/")
}

func (s *ZkDiscoveryService) createFullpath() error {
	for i := 1; i <= len(s.path); i++ {
		newpath := "/" + strings.Join(s.path[:i], "/")
		_, err := s.conn.Create(newpath, []byte{1}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			// It's OK if key already existed. Just skip.
			if err != zk.ErrNodeExists {
				return err
			}
		}
	}
	return nil
}

func (s *ZkDiscoveryService) Initialize(uris string, heartbeat int) error {
	var (
		// split here because uris can contain multiples ips
		// like `zk://192.168.0.1,192.168.0.2,192.168.0.3/path`
		parts = strings.SplitN(uris, "/", 2)
		ips   = strings.Split(parts[0], ",")
	)

	if len(parts) != 2 {
		return fmt.Errorf("invalid format %q, missing <path>", uris)
	}

	if strings.Contains(parts[1], "/") {
		s.path = strings.Split(parts[1], "/")
	} else {
		s.path = []string{parts[1]}
	}

	conn, _, err := zk.Connect(ips, time.Second)
	if err != nil {
		return err
	}

	s.conn = conn
	s.heartbeat = heartbeat
	err = s.createFullpath()
	if err != nil {
		return err
	}

	return nil
}

func (s *ZkDiscoveryService) Fetch() ([]*discovery.Node, error) {
	addrs, _, err := s.conn.Children(s.fullpath())

	if err != nil {
		return nil, err
	}

	return s.createNodes(addrs)
}

func (s *ZkDiscoveryService) createNodes(addrs []string) ([]*discovery.Node, error) {
	nodes := make([]*discovery.Node, 0)
	if addrs == nil {
		return nodes, nil
	}

	for _, addr := range addrs {
		node, err := discovery.NewNode(addr)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (s *ZkDiscoveryService) Watch(callback discovery.WatchCallback) {

	addrs, _, eventChan, err := s.conn.ChildrenW(s.fullpath())
	if err != nil {
		log.WithField("name", "zk").Debug("Discovery watch aborted")
		return
	}
	nodes, err := s.createNodes(addrs)
	if err == nil {
		callback(nodes)
	}

	for e := range eventChan {
		if e.Type == zk.EventNodeChildrenChanged {
			log.WithField("name", "zk").Debug("Discovery watch triggered")
			nodes, err := s.Fetch()
			if err == nil {
				callback(nodes)
			}
		}

	}

}

func (s *ZkDiscoveryService) Register(addr string) error {
	nodePath := path.Join(s.fullpath(), addr)

	// check existing for the parent path first
	exist, _, err := s.conn.Exists(s.fullpath())
	if err != nil {
		return err
	}

	// if the parent path does not exist yet
	if exist == false {
		// create the parent first
		err = s.createFullpath()
		if err != nil {
			return err
		}
	} else {
		// if node path exists
		exist, _, err = s.conn.Exists(nodePath)
		if err != nil {
			return err
		}
		// delete it first
		if exist {
			err = s.conn.Delete(nodePath, -1)
			if err != nil {
				return err
			}
		}
	}

	// create the node path to store address information
	_, err = s.conn.Create(nodePath, []byte(addr), 0, zk.WorldACL(zk.PermAll))
	return err
}
