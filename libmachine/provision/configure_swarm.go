package provision

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcndockerclient"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/samalba/dockerclient"
)

func configureSwarm(p Provisioner, swarmOptions swarm.Options, authOptions auth.Options) error {
	if !swarmOptions.IsSwarm {
		return nil
	}

	log.Info("Configuring swarm...")

	ip, err := p.GetDriver().GetIP()
	if err != nil {
		return err
	}

	u, err := url.Parse(swarmOptions.Host)
	if err != nil {
		return err
	}

	parts := strings.Split(u.Host, ":")
	port := parts[1]

	dockerPort := "2376"
	dockerDir := p.GetDockerOptionsDir()
	dockerHost := fmt.Sprintf("tcp://%s:%s", ip, dockerPort)
	dockerClient := mcndockerclient.RemoteDocker{dockerHost, &authOptions}
	advertiseInfo := fmt.Sprintf("%s:%s", ip, dockerPort)

	log.Info("SwarmOptions.Env : ", swarmOptions.Env)

	if swarmOptions.Master {
		cmd := fmt.Sprintf("manage --tlsverify --tlscacert=%s --tlscert=%s --tlskey=%s -H %s --strategy %s --advertise %s",
			authOptions.CaCertRemotePath,
			authOptions.ServerCertRemotePath,
			authOptions.ServerKeyRemotePath,
			swarmOptions.Host,
			swarmOptions.Strategy,
			advertiseInfo,
		)

		cmdMaster := strings.Fields(cmd)
		for _, option := range swarmOptions.ArbitraryFlags {
			cmdMaster = append(cmdMaster, "--"+option)
		}

		//Discovery must be at end of command
		cmdMaster = append(cmdMaster, swarmOptions.Discovery)

		hostBind := fmt.Sprintf("%s:%s", dockerDir, dockerDir)
		masterHostConfig := dockerclient.HostConfig{
			RestartPolicy: dockerclient.RestartPolicy{
				Name:              "Always",
				MaximumRetryCount: 0,
			},
			Binds:        []string{hostBind},
			PortBindings: map[string][]dockerclient.PortBinding{"3376/tcp": {{"", port}}},
			NetworkMode:  "host",
		}

		swarmMasterConfig := &dockerclient.ContainerConfig{
			Image: swarmOptions.Image,
			Env:   swarmOptions.Env,
			ExposedPorts: map[string]struct{}{
				"2375/tcp": {},
				"3376/tcp": {},
			},
			Cmd:        cmdMaster,
			HostConfig: masterHostConfig,
		}

		//Check if container "swarm-agent-master" already present, just start and break
		id, err := FindContainer(dockerClient, "swarm-agent-master")
		if err != nil {
			if err = CreateContainer(dockerClient, swarmMasterConfig, "swarm-agent-master"); err != nil {
				return err
			}
		} else {
			if err = StartContainer(dockerClient, id, &masterHostConfig); err != nil {
				return err
			}
		}
	}

	workerHostConfig := dockerclient.HostConfig{
		RestartPolicy: dockerclient.RestartPolicy{
			Name:              "Always",
			MaximumRetryCount: 0,
		},
		NetworkMode: "host",
	}

	swarmWorkerConfig := &dockerclient.ContainerConfig{
		Image: swarmOptions.Image,
		Env:   swarmOptions.Env,
		Cmd: []string{
			"join",
			"--advertise",
			advertiseInfo,
			swarmOptions.Discovery,
		},
		HostConfig: workerHostConfig,
	}

	//Check if, "swarm-agent" present, just start and skip creation.
	id, err := FindContainer(dockerClient, "swarm-agent")
	if err != nil {
		if err = CreateContainer(dockerClient, swarmWorkerConfig, "swarm-agent"); err != nil {
			return err
		}
	} else {
		if err = StartContainer(dockerClient, id, &workerHostConfig); err != nil {
			return err
		}
	}

	return nil
}
