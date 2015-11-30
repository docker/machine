package commands

import (
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/stretchr/testify/assert"
)

func TestCmdRmMissingMachineName(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{}
	api := &libmachinetest.FakeAPI{}

	err := cmdRm(commandLine, api)

	assert.EqualError(t, err, "Error: Expected to get one or more machine names as arguments")
	assert.True(t, commandLine.HelpShown)
}

func TestCmdRm(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1", "machineToRemove2"},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &fakedriver.Driver{},
			},
			{
				Name:   "machineToRemove2",
				Driver: &fakedriver.Driver{},
			},
			{
				Name:   "machine",
				Driver: &fakedriver.Driver{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.NoError(t, err)

	assert.False(t, libmachinetest.Exists(api, "machineToRemove1"))
	assert.False(t, libmachinetest.Exists(api, "machineToRemove2"))
	assert.True(t, libmachinetest.Exists(api, "machine"))
}
