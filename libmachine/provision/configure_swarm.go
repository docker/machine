package provision

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
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

	enginePort := engine.DefaultPort
	engineURL, err := p.GetDriver().GetURL()
	if err != nil {
		return err
	}

	parts := strings.Split(engineURL, ":")
	if len(parts) == 3 {
		dPort, err := strconv.Atoi(parts[2])
		if err != nil {
			return err
		}
		enginePort = dPort
	}

	parts = strings.Split(u.Host, ":")
	port := parts[1]

	dockerDir := p.GetDockerOptionsDir()
	dockerHost := &mcndockerclient.RemoteDocker{
		HostURL:    fmt.Sprintf("tcp://%s:%d", ip, enginePort),
		AuthOption: &authOptions,
	}
	advertiseInfo := fmt.Sprintf("%s:%d", ip, enginePort)

	if swarmOptions.Master {
		advertiseMasterInfo := fmt.Sprintf("%s:%s", ip, "3376")
		cmd := fmt.Sprintf("manage --tlsverify --tlscacert=%s --tlscert=%s --tlskey=%s -H %s --strategy %s --advertise %s",
			authOptions.CaCertRemotePath,
			authOptions.ServerCertRemotePath,
			authOptions.ServerKeyRemotePath,
			swarmOptions.Host,
			swarmOptions.Strategy,
			advertiseMasterInfo,
		)
		if swarmOptions.IsExperimental {
			cmd = "--experimental " + cmd
		}

		cmdMaster := strings.Fields(cmd)
		for _, option := range swarmOptions.ArbitraryFlags {
			cmdMaster = append(cmdMaster, "--"+option)
		}

		//Discovery must be at end of command
		cmdMaster = append(cmdMaster, swarmOptions.Discovery)

		hostBind := fmt.Sprintf("%s:%s", dockerDir, dockerDir)
		masterHostConfig := dockerclient.HostConfig{
			RestartPolicy: dockerclient.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
			Binds: []string{hostBind},
			PortBindings: map[string][]dockerclient.PortBinding{
				"3376/tcp": {
					{
						HostIp:   "0.0.0.0",
						HostPort: port,
					},
				},
			},
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

		err = mcndockerclient.CreateContainer(dockerHost, swarmMasterConfig, "swarm-agent-master")
		if err != nil {
			return err
		}
	}

	workerHostConfig := dockerclient.HostConfig{
		RestartPolicy: dockerclient.RestartPolicy{
			Name:              "always",
			MaximumRetryCount: 0,
		},
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
	if swarmOptions.IsExperimental {
		swarmWorkerConfig.Cmd = append([]string{"--experimental"}, swarmWorkerConfig.Cmd...)
	}

	return mcndockerclient.CreateContainer(dockerHost, swarmWorkerConfig, "swarm-agent")
}
