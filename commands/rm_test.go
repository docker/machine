package commands

import (
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/host"
	"github.com/stretchr/testify/assert"
)

func TestCmdRmMissingMachineName(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{}
	api := &commandstest.FakeLibmachineAPI{}

	err := cmdRm(commandLine, api)

	assert.EqualError(t, err, "Error: Expected to get one or more machine names as arguments")
	assert.True(t, commandLine.HelpShown)
}

func TestCmdRm(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1", "machineToRemove2"},
	}
	api := &commandstest.FakeLibmachineAPI{
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

	assert.False(t, commandstest.Exists(api, "machineToRemove1"))
	assert.False(t, commandstest.Exists(api, "machineToRemove2"))
	assert.True(t, commandstest.Exists(api, "machine"))
}
