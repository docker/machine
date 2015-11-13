package commands

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/hosttest"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

var (
	stdout *os.File
)

func init() {
	stdout = os.Stdout
}

func cleanup() {
	os.Stdout = stdout
}

func TestRunActionForeachMachine(t *testing.T) {
	// Assume a bunch of machines in randomly started or
	// stopped states.
	machines := []*host.Host{
		{
			Name:       "foo",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
			},
		},
		{
			Name:       "bar",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name: "baz",
			// Ssh, don't tell anyone but this
			// driver only _thinks_ it's named
			// virtualbox...  (to test serial actions)
			// It's actually FakeDriver!
			DriverName: "virtualbox",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name:       "spam",
			DriverName: "virtualbox",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
			},
		},
		{
			Name:       "eggs",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name:       "ham",
			DriverName: "fakedriver",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
			},
		},
	}

	runActionForeachMachine(func(h *host.Host) error {
		return h.Start()
	}, machines)

	for _, machine := range machines {
		machineState, _ := machine.Driver.GetState()

		assert.Equal(t, state.Running, machineState)
	}

	runActionForeachMachine(func(h *host.Host) error {
		return h.Stop()
	}, machines)

	for _, machine := range machines {
		machineState, _ := machine.Driver.GetState()

		assert.Equal(t, state.Stopped, machineState)
	}
}

func TestPrintIPEmptyGivenLocalEngine(t *testing.T) {
	defer cleanup()
	host, _ := hosttest.GetDefaultTestHost()

	out, w := captureStdout()

	assert.Nil(t, host.PrintIP())
	w.Close()

	assert.Equal(t, "", strings.TrimSpace(<-out))
}

func TestPrintIPPrintsGivenRemoteEngine(t *testing.T) {
	defer cleanup()
	host, _ := hosttest.GetDefaultTestHost()
	host.Driver = &fakedriver.Driver{}

	out, w := captureStdout()

	assert.Nil(t, host.PrintIP())
	w.Close()

	assert.Equal(t, "1.2.3.4", strings.TrimSpace(<-out))
}

func TestConsolidateError(t *testing.T) {
	cases := []struct {
		inputErrs   []error
		expectedErr error
	}{
		{
			inputErrs: []error{
				errors.New("Couldn't remove host 'bar'"),
			},
			expectedErr: errors.New("Couldn't remove host 'bar'"),
		},
		{
			inputErrs: []error{
				errors.New("Couldn't remove host 'bar'"),
				errors.New("Couldn't remove host 'foo'"),
			},
			expectedErr: errors.New("Couldn't remove host 'bar'\nCouldn't remove host 'foo'"),
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.expectedErr, consolidateErrs(c.inputErrs))
	}
}
