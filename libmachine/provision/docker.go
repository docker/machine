package provision

import (
	"fmt"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcndockerclient"
	"github.com/samalba/dockerclient"
)

// DockerClient implements DockerHost(mcndockerclient) interface
type DockerClient struct {
	HostURL    string
	AuthOption auth.Options
}

// URL returns the Docker host URL
func (dc DockerClient) URL() (string, error) {
	if dc.HostURL == "" {
		return "", fmt.Errorf("Docker Host URL not set")
	}

	return dc.HostURL, nil
}

// AuthOptions returns the authOptions
func (dc DockerClient) AuthOptions() *auth.Options {
	return &dc.AuthOption
}

//CreateContainer creates a docker container.
func CreateContainer(dockerHost DockerClient, config *dockerclient.ContainerConfig, name string) error {

	docker, err := mcndockerclient.DockerClient(dockerHost)
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

//FindContainer finds the container on docker host
func FindContainer(dockerHost DockerClient, name string) (string, error) {

	docker, err := mcndockerclient.DockerClient(dockerHost)
	if err != nil {
		return "", fmt.Errorf("Unable to connect to docker Client: %s", err)
	}

	containers, err := docker.ListContainers(true, true, "")
	if err != nil {
		return "", fmt.Errorf("No swarm containers found: %s", err)
	}

	for _, container := range containers {
		for _, c := range container.Names {
			log.Info("Name: ", c)
			if c == "/"+name {
				return container.Id, nil
			}
		}
	}

	return "", fmt.Errorf("Container : %s not found", name)
}

//StartContainer ..
func StartContainer(dockerHost DockerClient, id string, config *dockerclient.HostConfig) error {

	docker, err := mcndockerclient.DockerClient(dockerHost)
	if err != nil {
		return err
	}

	if err := docker.StartContainer(id, config); err != nil {
		return fmt.Errorf("Error while starting container: %s", err)
	}

	return nil
}
