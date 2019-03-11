package commands

import (
	"os/exec"
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/stretchr/testify/assert"
)

func TestCmdAddCACertificate(t *testing.T) {
	testCases := []struct {
		commandLine        CommandLine
		api                libmachine.API
		expectedErr        error
		sshClientCreator   host.SSHClientCreator
		expectedSSHShell   []string
		assertOnScpCommand func(cmd exec.Cmd) bool
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name:   "foo",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
				},
			},
			expectedErr:      errWrongNumberArguments,
			sshClientCreator: &FakeSSHClientCreator{},
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"somefile.txt"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name:   "foo",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
				},
			},
			expectedErr:      errFirstArgIsNotPEMFile,
			sshClientCreator: &FakeSSHClientCreator{},
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"somefile.pem"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name:   "foo",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
				},
			},
			expectedErr:      ErrNoDefault,
			sshClientCreator: &FakeSSHClientCreator{},
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"somefile.pem", "default"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name:   "default",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
				},
			},
			expectedErr:      errStateInvalidForSSH{"default"},
			sshClientCreator: &FakeSSHClientCreator{},
		},
		// This one can timeout hence the comment
		// {
		// 	commandLine: &commandstest.FakeCommandLine{
		// 		CliArgs: []string{"somefile.pem", "default"},
		// 		LocalFlags: &commandstest.FakeFlagger{
		// 			Data: map[string]interface{}{
		// 				"restart": true,
		// 			},
		// 		},
		// 	},
		// 	api: &libmachinetest.FakeAPI{
		// 		Hosts: []*host.Host{
		// 			{
		// 				Name: "default",
		// 				Driver: &fakedriver.Driver{
		// 					MockState: state.Running,
		// 				},
		// 				HostOptions: &host.Options{
		// 					EngineOptions: &engine.Options{},
		// 					AuthOptions:   &auth.Options{},
		// 					SwarmOptions:  &swarm.Options{},
		// 				},
		// 			},
		// 		},
		// 	},
		// 	expectedErr:      errStateInvalidForSSH{"default"},
		// 	sshClientCreator: &FakeSSHClientCreator{},
		// },
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"somefile.pem", "default"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "default",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
						},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
				},
			},
			expectedErr:      nil,
			sshClientCreator: &FakeSSHClientCreator{},
			assertOnScpCommand: func(cmd exec.Cmd) bool {
				lastTwoArgs := cmd.Args[len(cmd.Args)-2:]
				assert.Equal(t, "somefile.pem", lastTwoArgs[0])
				assert.Equal(t, "@:/var/lib/boot2docker/certs/", lastTwoArgs[1])
				return true
			},
		},
	}

	for _, tc := range testCases {
		host.SetSSHClientCreator(tc.sshClientCreator)

		SetCommandRunner(&MockCommandRunner{
			assertion: tc.assertOnScpCommand,
		})
		err := cmdAddCACertificate(tc.commandLine, tc.api)

		assert.Equal(t, tc.expectedErr, err)

	}
}
