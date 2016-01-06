package client

import (
	"github.com/docker/engine-api/types/container"
)

// ContainerUpdate updates resources of a container
func (cli *Client) ContainerUpdate(containerID string, hostConfig container.HostConfig) error {
	resp, err := cli.post("/containers/"+containerID+"/update", nil, hostConfig, nil)
	ensureReaderClosed(resp)
	return err
}
