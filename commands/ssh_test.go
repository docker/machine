package commands

import (
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/stretchr/testify/assert"
)

func TestCmdSSH(t *testing.T) {
	testCases := []struct {
		commandLine CommandLine
		api         libmachine.API
		expectedErr error
		helpShown   bool
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"-h"},
			},
			api:         &libmachinetest.FakeAPI{},
			expectedErr: nil,
			helpShown:   true,
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"--help"},
			},
			api:         &libmachinetest.FakeAPI{},
			expectedErr: nil,
			helpShown:   true,
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{""},
			},
			api:         &libmachinetest.FakeAPI{},
			expectedErr: ErrExpectedOneMachine,
		},
	}

	for _, tc := range testCases {
		err := cmdSSH(tc.commandLine, tc.api)
		assert.Equal(t, err, tc.expectedErr)

		if fcl, ok := tc.commandLine.(*commandstest.FakeCommandLine); ok {
			assert.Equal(t, tc.helpShown, fcl.HelpShown)
		}
	}
}
