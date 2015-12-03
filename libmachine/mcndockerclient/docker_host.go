package mcndockerclient

import "github.com/docker/machine/libmachine/auth"

type URLer interface {
	URL() (string, error)
}

type AuthOptionser interface {
	AuthOptions() *auth.Options
}

type DockerHost interface {
	URLer
	AuthOptionser
}
