package mcndockerclient

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerClient creates a docker client for a given host.
func DockerClient(dockerHost DockerHost) (*client.Client, error) {
	url, err := dockerHost.URL()
	if err != nil {
		return nil, err
	}

	authOptions := dockerHost.AuthOptions()
	return client.NewClientWithOpts(client.WithHost(url), client.WithTLSClientConfig(authOptions.CaCertPath, authOptions.ClientCertPath, authOptions.ClientKeyPath))
}

// ContainerConfig contains options needed to create and start a container
type ContainerConfig struct {
	Image        string
	Env          []string
	ExposedPorts map[string]struct{}
	Cmd          []string
	HostConfig   *container.HostConfig
}

// CreateContainer creates a docker container.
func CreateContainer(ctx context.Context, dockerHost DockerHost, config *ContainerConfig, name string) error {
	docker, err := DockerClient(dockerHost)
	if err != nil {
		return err
	}

	rc, err := docker.ImagePull(ctx, config.Image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Unable to pull image: %s", err)
	}

	_, err = io.Copy(ioutil.Discard, rc)
	if err != nil {
		return fmt.Errorf("Unable to read image pull status: %s", err)
	}

	err = rc.Close()
	if err != nil {
		return fmt.Errorf("Unable to close image pull status: %s", err)
	}

	containerConfig := &container.Config{
		Env:          config.Env,
		Cmd:          strslice.StrSlice(config.Cmd),
		ExposedPorts: map[nat.Port]struct{}{},
	}
	for k := range config.ExposedPorts {
		containerConfig.ExposedPorts[nat.Port(k)] = struct{}{}
	}

	createdContainer, err := docker.ContainerCreate(ctx, containerConfig, config.HostConfig, nil, name)
	if err != nil {
		return fmt.Errorf("Error while creating container: %s", err)
	}

	if err = docker.ContainerStart(ctx, createdContainer.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Error while starting container: %s", err)
	}

	return nil
}
