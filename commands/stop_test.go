package commands

import (
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestCmdStopMissingMachineName(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{}
	api := &libmachinetest.FakeAPI{}

	err := cmdStop(commandLine, api)

	assert.EqualError(t, err, "Error: Expected to get one or more machine names as arguments")
}

func TestCmdStop(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToStop1", "machineToStop2"},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name: "machineToStop1",
				Driver: &fakedriver.Driver{
					MockState: state.Running,
				},
			},
			{
				Name: "machineToStop2",
				Driver: &fakedriver.Driver{
					MockState: state.Running,
				},
			},
			{
				Name: "machine",
				Driver: &fakedriver.Driver{
					MockState: state.Running,
				},
			},
		},
	}

	err := cmdStop(commandLine, api)
	assert.NoError(t, err)

	assert.Equal(t, state.Stopped, libmachinetest.State(api, "machineToStop1"))
	assert.Equal(t, state.Stopped, libmachinetest.State(api, "machineToStop2"))
	assert.Equal(t, state.Running, libmachinetest.State(api, "machine"))
}
