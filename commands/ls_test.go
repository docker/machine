package commands

import (
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/state"
	"github.com/stretchr/testify/assert"
)

func TestParseFiltersErrorsGivenInvalidFilter(t *testing.T) {
	_, err := parseFilters([]string{"foo=bar"})
	assert.EqualError(t, err, "Unsupported filter key 'foo'")
}

func TestParseFiltersSwarm(t *testing.T) {
	actual, _ := parseFilters([]string{"swarm=foo"})
	assert.Equal(t, actual, FilterOptions{SwarmName: []string{"foo"}})
}

func TestParseFiltersDriver(t *testing.T) {
	actual, _ := parseFilters([]string{"driver=bar"})
	assert.Equal(t, actual, FilterOptions{DriverName: []string{"bar"}})
}

func TestParseFiltersState(t *testing.T) {
	actual, _ := parseFilters([]string{"state=Running"})
	assert.Equal(t, actual, FilterOptions{State: []string{"Running"}})
}

func TestParseFiltersAll(t *testing.T) {
	actual, _ := parseFilters([]string{"swarm=foo", "driver=bar", "state=Stopped"})
	assert.Equal(t, actual, FilterOptions{SwarmName: []string{"foo"}, DriverName: []string{"bar"}, State: []string{"Stopped"}})
}

func TestParseFiltersDuplicates(t *testing.T) {
	actual, _ := parseFilters([]string{"swarm=foo", "driver=bar", "swarm=baz", "driver=qux", "state=Running", "state=Starting"})
	assert.Equal(t, actual, FilterOptions{SwarmName: []string{"foo", "baz"}, DriverName: []string{"bar", "qux"}, State: []string{"Running", "Starting"}})
}

func TestParseFiltersValueWithEqual(t *testing.T) {
	actual, _ := parseFilters([]string{"driver=bar=baz"})
	assert.Equal(t, actual, FilterOptions{DriverName: []string{"bar=baz"}})
}

func TestFilterHostsReturnsSameGivenNoFilters(t *testing.T) {
	opts := FilterOptions{}
	hosts := []*libmachine.Host{
		{
			Name:        "testhost",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
		},
	}
	actual := filterHosts(hosts, opts)
	assert.EqualValues(t, actual, hosts)
}

func TestFilterHostsReturnsEmptyGivenEmptyHosts(t *testing.T) {
	opts := FilterOptions{
		SwarmName: []string{"foo"},
	}
	hosts := []*libmachine.Host{}
	assert.Empty(t, filterHosts(hosts, opts))
}

func TestFilterHostsReturnsEmptyGivenNonMatchingFilters(t *testing.T) {
	opts := FilterOptions{
		SwarmName: []string{"foo"},
	}
	hosts := []*libmachine.Host{
		{
			Name:        "testhost",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
		},
	}
	assert.Empty(t, filterHosts(hosts, opts))
}

func TestFilterHostsBySwarmName(t *testing.T) {
	opts := FilterOptions{
		SwarmName: []string{"master"},
	}
	master :=
		&libmachine.Host{
			Name: "master",
			HostOptions: &libmachine.HostOptions{
				SwarmOptions: &swarm.SwarmOptions{Master: true, Discovery: "foo"},
			},
		}
	node1 :=
		&libmachine.Host{
			Name: "node1",
			HostOptions: &libmachine.HostOptions{
				SwarmOptions: &swarm.SwarmOptions{Master: false, Discovery: "foo"},
			},
		}
	othermaster :=
		&libmachine.Host{
			Name: "othermaster",
			HostOptions: &libmachine.HostOptions{
				SwarmOptions: &swarm.SwarmOptions{Master: true, Discovery: "bar"},
			},
		}
	hosts := []*libmachine.Host{master, node1, othermaster}
	expected := []*libmachine.Host{master, node1}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsByDriverName(t *testing.T) {
	opts := FilterOptions{
		DriverName: []string{"fakedriver"},
	}
	node1 :=
		&libmachine.Host{
			Name:        "node1",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
		}
	node2 :=
		&libmachine.Host{
			Name:        "node2",
			DriverName:  "virtualbox",
			HostOptions: &libmachine.HostOptions{},
		}
	node3 :=
		&libmachine.Host{
			Name:        "node3",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
		}
	hosts := []*libmachine.Host{node1, node2, node3}
	expected := []*libmachine.Host{node1, node3}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsByState(t *testing.T) {
	opts := FilterOptions{
		State: []string{"Paused", "Saved", "Stopped"},
	}
	node1 :=
		&libmachine.Host{
			Name:        "node1",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
			Driver:      &fakedriver.FakeDriver{MockState: state.Paused},
		}
	node2 :=
		&libmachine.Host{
			Name:        "node2",
			DriverName:  "virtualbox",
			HostOptions: &libmachine.HostOptions{},
			Driver:      &fakedriver.FakeDriver{MockState: state.Stopped},
		}
	node3 :=
		&libmachine.Host{
			Name:        "node3",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
			Driver:      &fakedriver.FakeDriver{MockState: state.Running},
		}
	hosts := []*libmachine.Host{node1, node2, node3}
	expected := []*libmachine.Host{node1, node2}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsMultiFlags(t *testing.T) {
	opts := FilterOptions{
		SwarmName:  []string{},
		DriverName: []string{"fakedriver", "virtualbox"},
	}
	node1 :=
		&libmachine.Host{
			Name:        "node1",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
		}
	node2 :=
		&libmachine.Host{
			Name:        "node2",
			DriverName:  "virtualbox",
			HostOptions: &libmachine.HostOptions{},
		}
	node3 :=
		&libmachine.Host{
			Name:        "node3",
			DriverName:  "softlayer",
			HostOptions: &libmachine.HostOptions{},
		}
	hosts := []*libmachine.Host{node1, node2, node3}
	expected := []*libmachine.Host{node1, node2}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsDifferentFlagsProduceAND(t *testing.T) {
	opts := FilterOptions{
		DriverName: []string{"virtualbox"},
		State:      []string{"Running"},
	}
	node1 :=
		&libmachine.Host{
			Name:        "node1",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
			Driver:      &fakedriver.FakeDriver{MockState: state.Paused},
		}
	node2 :=
		&libmachine.Host{
			Name:        "node2",
			DriverName:  "virtualbox",
			HostOptions: &libmachine.HostOptions{},
			Driver:      &fakedriver.FakeDriver{MockState: state.Stopped},
		}
	node3 :=
		&libmachine.Host{
			Name:        "node3",
			DriverName:  "fakedriver",
			HostOptions: &libmachine.HostOptions{},
			Driver:      &fakedriver.FakeDriver{MockState: state.Running},
		}
	hosts := []*libmachine.Host{node1, node2, node3}
	expected := []*libmachine.Host{}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}
