package provision

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcndockerclient"
	"github.com/docker/machine/libmachine/swarm"
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
	dockerHost := &mcndockerclient.RemoteDocker{
		HostURL:    fmt.Sprintf("tcp://%s:%s", ip, dockerPort),
		AuthOption: &authOptions,
	}
	advertiseInfo := fmt.Sprintf("%s:%s", ip, dockerPort)

	if swarmOptions.Master {
		err = configureSwarmMaster(dockerHost, swarmOptions, authOptions, dockerDir, port, advertiseInfo)
		if err != nil {
			return err
		}
	}

	return configureSwarmAgent(dockerHost, swarmOptions, advertiseInfo)
}

func configureSwarmMaster(dockerHost mcndockerclient.DockerHost, swarmOptions swarm.Options, authOptions auth.Options, dockerDir string, port string, advertiseInfo string) error {
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
	// Discovery must be at end of command
	cmdMaster = append(cmdMaster, swarmOptions.Discovery)

	exposedPorts, _, err := nat.ParsePortSpecs([]string{"2375/tcp", "3376/tcp"})
	if err != nil {
		return err
	}

	swarmMasterConfig := &container.Config{
		Image:        swarmOptions.Image,
		Env:          swarmOptions.Env,
		ExposedPorts: exposedPorts,
		Cmd:          strslice.New(cmdMaster...),
	}

	_, portBindings, err := nat.ParsePortSpecs([]string{fmt.Sprintf("%s:3376/tcp", port)})
	if err != nil {
		return err
	}

	hostBind := fmt.Sprintf("%s:%s", dockerDir, dockerDir)
	masterHostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Binds:        []string{hostBind},
		PortBindings: portBindings,
		NetworkMode:  "host",
	}

	return mcndockerclient.CreateContainer(dockerHost, swarmMasterConfig, masterHostConfig, "swarm-agent-master")
}

func configureSwarmAgent(dockerHost mcndockerclient.DockerHost, swarmOptions swarm.Options, advertiseInfo string) error {
	swarmWorkerConfig := &container.Config{
		Image: swarmOptions.Image,
		Env:   swarmOptions.Env,
		Cmd: strslice.New(
			"join",
			"--advertise",
			advertiseInfo,
			swarmOptions.Discovery,
		),
	}

	workerHostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		NetworkMode: "host",
	}

	return mcndockerclient.CreateContainer(dockerHost, swarmWorkerConfig, workerHostConfig, "swarm-agent")
}
