package mcndockerclient

import (
	"fmt"

	"github.com/docker/machine/libmachine/cert"
	"github.com/samalba/dockerclient"
)

func DockerClient(host DockerHost) (*dockerclient.DockerClient, error) {
	url, err := host.URL()
	if err != nil {
		return nil, err
	}

	tlsConfig, err := cert.ReadTLSConfig(url, host.AuthOptions())
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLS config: %s", err)
	}

	return dockerclient.NewDockerClient(url, tlsConfig)
}

//CreateContainer creates a docker container.
func CreateContainer(dockerHost RemoteDocker, config *dockerclient.ContainerConfig, name string) error {
	docker, err := DockerClient(dockerHost)
	if err != nil {
		return err
	}

	if err = docker.PullImage(config.Image, nil); err != nil {
		return fmt.Errorf("Unable to Pull Image: %s", err)
	}

	containerID, err := docker.CreateContainer(config, name)
	if err != nil {
		return fmt.Errorf("Error while creating container: %s", err)
	}

	if err = docker.StartContainer(containerID, &config.HostConfig); err != nil {
		return fmt.Errorf("Error while starting container: %s", err)
	}

	return nil
}
