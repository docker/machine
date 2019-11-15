package commands

import (
	"testing"

	"github.com/rancher/machine/commands/commandstest"
	"github.com/rancher/machine/drivers/fakedriver"
	"github.com/rancher/machine/libmachine"
	"github.com/rancher/machine/libmachine/auth"
	"github.com/rancher/machine/libmachine/engine"
	"github.com/rancher/machine/libmachine/host"
	"github.com/rancher/machine/libmachine/libmachinetest"
	"github.com/rancher/machine/libmachine/provision"
	"github.com/rancher/machine/libmachine/swarm"
	"github.com/stretchr/testify/assert"
)

func TestCmdProvision(t *testing.T) {
	testCases := []struct {
		commandLine CommandLine
		api         libmachine.API
		expectedErr error
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"foo", "bar"},
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
					{
						Name:   "bar",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
				},
			},
			expectedErr: nil,
		},
	}

	provision.SetDetector(&provision.FakeDetector{
		Provisioner: provision.NewFakeProvisioner(nil),
	})

	// fakeprovisioner always returns "true" for compatible host, so we
	// just need to register it.
	provision.Register("fakeprovisioner", &provision.RegisteredProvisioner{
		New: provision.NewFakeProvisioner,
	})

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedErr, cmdProvision(tc.commandLine, tc.api))
	}
}
