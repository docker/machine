package mcndockerclient

import (
	"fmt"

	"net/http"

	"io"
	"io/ioutil"

	"strings"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/machine/libmachine/cert"
)

// DockerClient creates a docker client for a given host.
func DockerClient(dockerHost DockerHost) (*client.Client, error) {
	url, err := dockerHost.URL()
	if err != nil {
		return nil, err
	}

	tlsConfig, err := cert.ReadTLSConfig(url, dockerHost.AuthOptions())
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLS config: %s", err)
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	defaultHeaders := map[string]string{"User-Agent": "docker-machine"}

	return client.NewClient(url, "", transport, defaultHeaders)
}

// CreateContainer creates a docker container.
func CreateContainer(dockerHost DockerHost, config *container.Config, hostConfig *container.HostConfig, containerName string) error {
	docker, err := DockerClient(dockerHost)
	if err != nil {
		return err
	}

	options := parseImagePullOptions(config.Image)

	body, err := docker.ImagePull(options, func() (string, error) { return "", nil })
	if err != nil {
		return fmt.Errorf("Unable to pull image: %s", err)
	}

	defer body.Close()
	io.Copy(ioutil.Discard, body)

	response, err := docker.ContainerCreate(config, hostConfig, containerName)
	if err != nil {
		return fmt.Errorf("Error while creating container: %s", err)
	}

	if err = docker.ContainerStart(response.ID); err != nil {
		return fmt.Errorf("Error while starting container: %s", err)
	}

	return nil
}

func parseImagePullOptions(imageName string) types.ImagePullOptions {
	parts := strings.Split(imageName, ":")

	imageId := parts[0]

	tag := "latest"
	if len(parts) > 1 {
		tag = parts[1]
	}

	return types.ImagePullOptions{
		ImageID: imageId,
		Tag:     tag,
	}
}
