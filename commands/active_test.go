package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmdActiveNone(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:        "host1",
			ActiveHost:  false,
			ActiveSwarm: false,
		},
		{
			Name:        "host2",
			ActiveHost:  false,
			ActiveSwarm: false,
		},
		{
			Name:        "host3",
			ActiveHost:  false,
			ActiveSwarm: false,
		},
	}
	_, err := activeHost(hostListItems)
	assert.Equal(t, err, errNoActiveHost)
}

func TestCmdActiveHost(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:        "host1",
			ActiveHost:  false,
			ActiveSwarm: false,
		},
		{
			Name:        "host2",
			ActiveHost:  true,
			ActiveSwarm: false,
		},
		{
			Name:        "host3",
			ActiveHost:  false,
			ActiveSwarm: false,
		},
	}
	active, err := activeHost(hostListItems)
	assert.Equal(t, err, nil)
	assert.Equal(t, active.Name, "host2")
}

func TestCmdActiveSwarm(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:        "host1",
			ActiveHost:  false,
			ActiveSwarm: false,
		},
		{
			Name:        "host2",
			ActiveHost:  false,
			ActiveSwarm: false,
		},
		{
			Name:        "host3",
			ActiveHost:  false,
			ActiveSwarm: true,
		},
	}
	active, err := activeHost(hostListItems)
	assert.Equal(t, err, nil)
	assert.Equal(t, active.Name, "host3")
}
