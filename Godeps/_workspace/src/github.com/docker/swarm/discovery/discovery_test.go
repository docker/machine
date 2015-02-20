package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	node, err := NewNode("127.0.0.1:2375")
	assert.Equal(t, node.Host, "127.0.0.1")
	assert.Equal(t, node.Port, "2375")
	assert.NoError(t, err)

	_, err = NewNode("127.0.0.1")
	assert.Error(t, err)
}

func TestParse(t *testing.T) {
	scheme, uri := parse("127.0.0.1:2375")
	assert.Equal(t, scheme, "nodes")
	assert.Equal(t, uri, "127.0.0.1:2375")

	scheme, uri = parse("localhost:2375")
	assert.Equal(t, scheme, "nodes")
	assert.Equal(t, uri, "localhost:2375")

	scheme, uri = parse("scheme://127.0.0.1:2375")
	assert.Equal(t, scheme, "scheme")
	assert.Equal(t, uri, "127.0.0.1:2375")

	scheme, uri = parse("scheme://localhost:2375")
	assert.Equal(t, scheme, "scheme")
	assert.Equal(t, uri, "localhost:2375")

	scheme, uri = parse("")
	assert.Equal(t, scheme, "nodes")
	assert.Equal(t, uri, "")
}
