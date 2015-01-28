package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	discovery := &FileDiscoveryService{}
	discovery.Initialize("/path/to/file", 0)
	assert.Equal(t, discovery.path, "/path/to/file")
}

func TestRegister(t *testing.T) {
	discovery := &FileDiscoveryService{path: "/path/to/file"}
	assert.Error(t, discovery.Register("0.0.0.0"))
}
