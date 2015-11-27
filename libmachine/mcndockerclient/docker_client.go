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
