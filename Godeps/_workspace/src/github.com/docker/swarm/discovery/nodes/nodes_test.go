package nodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialise(t *testing.T) {
	discovery := &NodesDiscoveryService{}
	discovery.Initialize("1.1.1.1:1111,2.2.2.2:2222", 0)
	assert.Equal(t, len(discovery.nodes), 2)
	assert.Equal(t, discovery.nodes[0].String(), "1.1.1.1:1111")
	assert.Equal(t, discovery.nodes[1].String(), "2.2.2.2:2222")
}

func TestRegister(t *testing.T) {
	discovery := &NodesDiscoveryService{}
	assert.Error(t, discovery.Register("0.0.0.0"))
}
