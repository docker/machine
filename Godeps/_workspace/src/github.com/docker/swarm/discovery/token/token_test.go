package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	discovery := &TokenDiscoveryService{}
	discovery.Initialize("token", 0)
	assert.Equal(t, discovery.token, "token")
	assert.Equal(t, discovery.url, DISCOVERY_URL)

	discovery.Initialize("custom/path/token", 0)
	assert.Equal(t, discovery.token, "token")
	assert.Equal(t, discovery.url, "https://custom/path")
}

func TestRegister(t *testing.T) {
	discovery := &TokenDiscoveryService{token: "TEST_TOKEN", url: DISCOVERY_URL}
	expected := "127.0.0.1:2675"
	assert.NoError(t, discovery.Register(expected))

	addrs, err := discovery.Fetch()
	assert.NoError(t, err)
	assert.Equal(t, len(addrs), 1)
	assert.Equal(t, addrs[0].String(), expected)

	assert.NoError(t, discovery.Register(expected))
}
